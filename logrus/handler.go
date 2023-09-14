package logrus

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"

	"github.com/sirupsen/logrus"

	"github.com/m40Jc001/slog-handler-adapter/helper"
)

/*
	implement log/slog.Handler
*/

var _ slog.Handler = (*Handler)(nil)

type Handler struct {
	logr      *logrus.Logger
	addSource bool
	isJSON    bool
	level     *slog.LevelVar
	attrGroup *helper.AttrGroup
}

type HandlerOptions struct {
	AddSource     bool
	JSONFormatter bool
	Level         slog.Level
}

func NewHandler(writer io.Writer, options *HandlerOptions) *Handler {
	logrus.New()
	logr := &logrus.Logger{
		Out:          writer,
		Hooks:        make(logrus.LevelHooks),
		Level:        logrus.TraceLevel, // control by "Enable" function
		ExitFunc:     os.Exit,
		ReportCaller: false, // always not use this, handle this feature at "Handle" function
	}
	if options.JSONFormatter {
		logr.Formatter = &logrus.JSONFormatter{DisableTimestamp: true}
	} else {
		logr.Formatter = &logrus.TextFormatter{DisableTimestamp: true}
	}

	levelar := &slog.LevelVar{}
	levelar.Set(options.Level)

	return &Handler{
		logr:      logr,
		addSource: options.AddSource,
		level:     levelar,
		attrGroup: &helper.AttrGroup{},
		isJSON:    options.JSONFormatter,
	}
}

func (h *Handler) clone() *Handler {
	return &Handler{
		logr:      h.logr,
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

	var fields logrus.Fields
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
		fields[timeKey] = r.Time
	}

	if h.addSource && r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		fields[fileKey] = fmt.Sprintf("%s:%d", f.File, f.Line)
		fields[funcKey] = f.Function
	}

	h.logr.WithFields(fields).Log(level2LogrusLevel(r.Level), r.Message)
	return nil
}

func attrs2TextLogrusField(attrs []slog.Attr) (m logrus.Fields, err error) {
	m = logrus.Fields{}
	var rec func(prefix string, attrs []slog.Attr) error
	rec = func(prefix string, attrs []slog.Attr) error {
		for _, attr := range attrs {
			if _, ok := m[prefix+attr.Key]; ok {
				return fmt.Errorf("dup key: %s", attr.Key)
			}

			if attr.Value.Kind() == slog.KindGroup {
				err = rec(prefix+attr.Key+".", attr.Value.Resolve().Group())
				if err != nil {
					return err
				}
			} else {
				m[prefix+attr.Key] = attr.Value.Any()
			}
		}
		return nil
	}
	return m, rec("", attrs)
}

func attrs2JSONLogrusField(attrs []slog.Attr) (m logrus.Fields, err error) {
	var rec func(m logrus.Fields, attrs []slog.Attr) error
	rec = func(m logrus.Fields, attrs []slog.Attr) error {
		for _, attr := range attrs {
			if _, ok := m[attr.Key]; ok {
				return fmt.Errorf("dup key: %s", attr.Key)
			}

			if attr.Value.Kind() == slog.KindGroup {
				inner := logrus.Fields{}
				err = rec(inner, attr.Value.Resolve().Group())
				if err != nil {
					return err
				}
				m[attr.Key] = inner
			} else {
				m[attr.Key] = attr.Value.Any()
			}
		}
		return nil
	}
	m = logrus.Fields{}
	err = rec(m, attrs)
	return m, err
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
