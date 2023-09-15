package zap

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/m40Jc001/slog-handler-adapter/helper"
)

/*
	implement log/slog.Handler
*/

var _ slog.Handler = (*Handler)(nil)

type Handler struct {
	core      zapcore.Core
	addSource bool
	isJSON    bool
	level     *slog.LevelVar
	attrGroup *helper.AttrGroup
}

type HandlerOptions struct {
	AddSource        bool
	JSONFormatter    bool
	Level            slog.Level
	EnableStacktrace bool
}

// NewHandler
// Taking into account that the concurrency safety of zap relies on the underlying writer's locking mechanism,
//
// here, during the creation of a NewHandler, a simple wrapper is used with zapcore.Lock and zapcore.AddSync.
//
// Of course, zapcore.Core is passed as a parameter to NewHandler, and I believe this is a viable approach.
func NewHandler(writer io.Writer, options *HandlerOptions) *Handler {
	var encoding string
	if options.JSONFormatter {
		encoding = "json"
	} else {
		encoding = "console"
	}
	cfg := zap.Config{
		Level:             zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Development:       false,
		DisableCaller:     true,
		DisableStacktrace: options.EnableStacktrace,
		Sampling:          nil,
		Encoding:          encoding,
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:          msgKey,
			LevelKey:            levelKey,
			TimeKey:             "",
			NameKey:             nameKey,
			CallerKey:           callerKey,
			FunctionKey:         funcKey,
			StacktraceKey:       stackKey,
			SkipLineEnding:      false,
			LineEnding:          "\n",
			EncodeLevel:         lowercaseLevelEncoderAddTraceLevel,
			EncodeTime:          nil,
			EncodeDuration:      zapcore.SecondsDurationEncoder,
			EncodeCaller:        zapcore.ShortCallerEncoder,
			EncodeName:          nil,
			NewReflectedEncoder: nil,
			ConsoleSeparator:    " ",
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		InitialFields:    map[string]interface{}{},
	}

	var encoder zapcore.Encoder
	if cfg.Encoding == "json" {
		encoder = zapcore.NewJSONEncoder(cfg.EncoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(cfg.EncoderConfig)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.Lock(zapcore.AddSync(writer)),
		zapcore.Level(cfg.Level.Level()),
	)

	levelar := &slog.LevelVar{}
	levelar.Set(options.Level)

	return &Handler{
		core:      core,
		addSource: options.AddSource,
		level:     levelar,
		attrGroup: &helper.AttrGroup{},
		isJSON:    options.JSONFormatter,
	}
}

func (h *Handler) clone() *Handler {
	return &Handler{
		core:      h.core.With([]zapcore.Field{}),
		addSource: h.addSource,
		level:     h.level,
		isJSON:    h.isJSON,
		attrGroup: h.attrGroup,
	}
}

// Enabled reports whether the handler handles records at the given level.
// The handler ignores records whose level is lower.
// It is called early, before any arguments are processed,
// to save effort if the log event should be discarded.
// If called from a Logger method, the first argument is the context
// passed to that method, or context.Background() if nil was passed
// or the method does not take a context.
// The context is passed so Enabled can use its values
// to make a decision.
func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	return h.level.Level() >= level
}

// Handle handles the Record.
// It will only be called when Enabled returns true.
// The Context argument is as for Enabled.
// It is present solely to provide Handlers access to the context's values.
// Canceling the context should not affect record processing.
// (Among other things, log messages may be necessary to debug a
// cancellation-related problem.)
//
// Handle methods that produce output should observe the following rules:
//   - If r.Time is the zero time, ignore the time.
//   - If r.PC is zero, ignore it.
//   - Attr's values should be resolved.
//   - If an Attr's key and value are both the zero value, ignore the Attr.
//     This can be tested with attr.Equal(Attr{}).
//   - If a group's key is empty, inline the group's Attrs.
//   - If a group has no Attrs (even if it has a non-empty key),
//     ignore it.
func (h *Handler) Handle(_ context.Context, r slog.Record) error {
	recordAttrs := []slog.Attr{}
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "" {
			return true
		}
		recordAttrs = append(recordAttrs, a)
		return true
	})

	var fields []zap.Field
	var err error

	if h.isJSON {
		fields, err = attrs2JSONLogrusField(h.attrGroup.WithAttrs(recordAttrs).Attrs())
	} else {
		fields, err = attrs2TextLogrusField(h.attrGroup.WithAttrs(recordAttrs).Attrs())
	}

	if err != nil {
		return err
	}

	if !r.Time.IsZero() {
		fields = append(fields, zap.Time(timeKey, r.Time))
	}

	if h.addSource && r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		fields = append(fields,
			zap.String(fileKey, fmt.Sprintf("%s:%d", f.File, f.Line)),
			zap.String(funcKey, f.Function),
		)
	}

	return h.core.Write(zapcore.Entry{
		Level:   level2ZapLevel(r.Level),
		Message: r.Message,
	}, fields)
}

func attrs2TextLogrusField(attrs []slog.Attr) (m []zap.Field, err error) {
	m = []zap.Field{}
	dupKeyMap := map[string]struct{}{}
	var rec func(prefix string, attrs []slog.Attr) error
	rec = func(prefix string, attrs []slog.Attr) error {
		for _, attr := range attrs {
			key := prefix + attr.Key
			if _, ok := dupKeyMap[key]; ok {
				return fmt.Errorf("dup key: %s", attr.Key)
			} else {
				dupKeyMap[key] = struct{}{}
			}

			if attr.Value.Kind() == slog.KindGroup {
				err = rec(prefix+attr.Key+".", attr.Value.Resolve().Group())
				if err != nil {
					return err
				}
			} else {
				m = append(m, zap.Any(key, attr.Value.Any()))
			}
		}
		return nil
	}
	return m, rec("", attrs)
}

func attrs2JSONLogrusField(attrs []slog.Attr) (m []zap.Field, err error) {
	dupKeyMap := map[string]struct{}{}
	for _, attr := range attrs {
		if _, ok := dupKeyMap[attr.Key]; ok {
			return nil, fmt.Errorf("dup key: %s", attr.Key)
		} else {
			dupKeyMap[attr.Key] = struct{}{}
		}

		if attr.Value.Kind() == slog.KindGroup {
			inner, err := attrs2JSONLogrusField(attr.Value.Resolve().Group())
			if err != nil {
				return nil, err
			}
			m = append(m, zap.Any(attr.Key, inner))
		} else {
			m = append(m, zap.Any(attr.Key, attr.Value.Any()))
		}
	}
	return
}

// WithAttrs returns a new Handler whose attributes consist of
// both the receiver's attributes and the arguments.
// The Handler owns the slice: it may retain, modify or discard it.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	cp := h.clone()
	cp.attrGroup = cp.attrGroup.WithAttrs(attrs)
	return cp
}

// WithGroup returns a new Handler with the given group appended to
// the receiver's existing groups.
// The keys of all subsequent attributes, whether added by With or in a
// Record, should be qualified by the sequence of group names.
//
// How this qualification happens is up to the Handler, so long as
// this Handler's attribute keys differ from those of another Handler
// with a different sequence of group names.
//
// A Handler should treat WithGroup as starting a Group of Attrs that ends
// at the end of the log event. That is,
//
//	logger.WithGroup("s").LogAttrs(level, msg, slog.Int("a", 1), slog.Int("b", 2))
//
// should behave like
//
//	logger.LogAttrs(level, msg, slog.Group("s", slog.Int("a", 1), slog.Int("b", 2)))
//
// If the name is empty, WithGroup returns the receiver.
func (h *Handler) WithGroup(name string) slog.Handler {
	cp := h.clone()
	cp.attrGroup = cp.attrGroup.WithGroup(name)
	return cp
}
