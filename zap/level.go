package zap

import (
	"log/slog"

	"go.uber.org/zap/zapcore"

	logger "github.com/m40Jc001/slog-handler-adapter"
)

const TraceLevel zapcore.Level = zapcore.DebugLevel - 1

func lowercaseLevelEncoderAddTraceLevel(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	if level == TraceLevel {
		enc.AppendString("trace")
		return
	}
	zapcore.LowercaseLevelEncoder(level, enc)
}

func level2ZapLevel(level slog.Level) zapcore.Level {
	switch level {
	case logger.LevelTrace:
		return TraceLevel
	case logger.LevelDebug:
		return zapcore.DebugLevel
	case logger.LevelInfo:
		return zapcore.InfoLevel
	case logger.LevelWarn:
		return zapcore.WarnLevel
	case logger.LevelError:
		return zapcore.ErrorLevel
	case logger.LevelPanic:
		return zapcore.PanicLevel
	case logger.LevelFatal:
		return zapcore.FatalLevel
	}
	return TraceLevel // TODO: Here, if the above does not map, it is converted to the lowest log level, not sure if there is a better way to do this
}
