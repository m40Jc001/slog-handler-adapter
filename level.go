package logger

import "log/slog"

const (
	LevelTrace = slog.Level(-8)  //
	LevelDebug = slog.LevelDebug // -4
	LevelInfo  = slog.LevelInfo  // 0
	LevelWarn  = slog.LevelWarn  // 4
	LevelError = slog.LevelError // 8
	LevelPanic = slog.Level(12)  // 12
	LevelFatal = slog.Level(16)
)
