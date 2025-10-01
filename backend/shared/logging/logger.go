package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

func New(serviceName string) *Logger {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	config.TimeKey = "timestamp"
	config.LevelKey = "level"
	config.NameKey = "logger"
	config.CallerKey = "caller"
	config.MessageKey = "message"
	config.StacktraceKey = "stacktrace"
	config.EncodeLevel = zapcore.LowercaseLevelEncoder
	config.EncodeCaller = zapcore.ShortCallerEncoder

	logLevel := os.Getenv("LOG_LEVEL")
	var level zapcore.Level
	switch logLevel {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}

	encoder := zapcore.NewJSONEncoder(config)
	defaultCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)

	logger := zap.New(defaultCore).With(zap.String("service", serviceName))

	return &Logger{logger}
}

func (l *Logger) WithFields(fields ...zap.Field) *Logger {
	return &Logger{l.Logger.With(fields...)}
}

func (l *Logger) Sync() error {
	return l.Logger.Sync()
}
