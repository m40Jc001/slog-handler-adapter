package logrus

import (
	"log/slog"

	"github.com/sirupsen/logrus"

	logger "github.com/m40Jc001/slog-handler-adapter"
)

func level2LogrusLevel(level slog.Level) logrus.Level {
	switch level {
	case logger.LevelTrace:
		return logrus.TraceLevel
	case logger.LevelDebug:
		return logrus.DebugLevel
	case logger.LevelInfo:
		return logrus.InfoLevel
	case logger.LevelWarn:
		return logrus.WarnLevel
	case logger.LevelError:
		return logrus.ErrorLevel
	case logger.LevelPanic:
		return logrus.PanicLevel
	case logger.LevelFatal:
		return logrus.FatalLevel
	}
	return logrus.TraceLevel // TODO: Here, if the above does not map, it is converted to the lowest log level, not sure if there is a better way to do this
}
