package logging

import (
	"fmt"
	"math"

	"go.uber.org/zap/zapcore"
)

const (
	// 禁用的日志记录级别，不会输出
	DisabledLevel = zapcore.Level(math.MinInt8)

	// 记录详细的消息调试信息
	PayloadLevel = zapcore.DebugLevel - 1
)

func NameToLevel(level string) zapcore.Level {
	l, err := nameToLevel(level)
	if err != nil {
		return zapcore.InfoLevel
	}
	return l
}

func nameToLevel(level string) (zapcore.Level, error) {
	switch level {
	case "PAYLOAD", "payload":
		return PayloadLevel, nil
	case "DEBUG", "debug":
		return zapcore.DebugLevel, nil
	case "INFO", "info":
		return zapcore.InfoLevel, nil
	case "WARNING", "WARN", "warning", "warn":
		return zapcore.WarnLevel, nil
	case "ERROR", "error":
		return zapcore.ErrorLevel, nil
	case "DPANIC", "dpanic":
		return zapcore.DPanicLevel, nil
	case "PANIC", "panic":
		return zapcore.FatalLevel, nil
	case "NOTICE", "notice":
		return zapcore.InfoLevel, nil
	case "CRITICAL", "critical":
		return zapcore.ErrorLevel, nil
	default:
		return DisabledLevel, fmt.Errorf("invalid log level: %s", level)
	}
}

func IsValidLevel(level string) bool {
	_, err := nameToLevel(level)
	return err == nil
}
