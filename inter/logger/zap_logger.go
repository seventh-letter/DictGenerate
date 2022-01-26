package logger

import (
	"go.uber.org/zap/zapcore"
	"strings"
)

type ZapLogger struct {
	Sync func() error
}

func NewZapLogger() *ZapLogger {
	return &ZapLogger{Sync: gLogger.Sync}
}

func ToZapLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case zapcore.DebugLevel.String():
		return zapcore.DebugLevel
	case zapcore.InfoLevel.String():
		return zapcore.InfoLevel
	case zapcore.WarnLevel.String():
		return zapcore.WarnLevel
	case zapcore.ErrorLevel.String():
		return zapcore.ErrorLevel
	case zapcore.FatalLevel.String():
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}
