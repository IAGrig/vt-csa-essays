package logging

import "go.uber.org/zap"

type EmptyLogger struct{}

func NewEmptyLogger() *Logger {
	zapLogger := zap.NewNop()
	return &Logger{Logger: zapLogger}
}

func (n *EmptyLogger) Debug(msg string, fields ...zap.Field) {}
func (n *EmptyLogger) Info(msg string, fields ...zap.Field)  {}
func (n *EmptyLogger) Warn(msg string, fields ...zap.Field)  {}
func (n *EmptyLogger) Error(msg string, fields ...zap.Field) {}
func (n *EmptyLogger) With(fields ...zap.Field) *Logger {
	return &Logger{Logger: zap.NewNop()}
}
func (n *EmptyLogger) Sync() error { return nil }
