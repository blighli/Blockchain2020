package logging

import (
	"strings"

	"go.uber.org/zap/zapcore"
)

const (
	// 默认输出日志级别，可通过FABRIC_LOGGING_SPEC环境变量修改
	defaultLevel  = zapcore.InfoLevel
	defaultFormat = "%{color}%{time:2006-01-02 15:04:05.000 MST} [%{module}] %{shortfunc} -> %{level:.4s} %{id:03x}%{color:reset} %{message}"
)

// 全局日志系统
var GlobalLogging *Logging

// init
func init() {
	logging, err := New(Config{})
	if err != nil {
		panic(err)
	}
	GlobalLogging = logging

	//grpcLogger := GlobalLogging.ZapLogger("grpc")
	//grpclog.SetLogger(NewGRPCLogger(grpcLogger))
}

// 根据配置初始化全局日志系统
func Init(config Config) {
	err := GlobalLogging.Apply(config)
	if err != nil {
		panic(err)
	}
}

// 根据日志记录器名获得日志记录级别
func GetLoggerLevel(loggerName string) string {
	return strings.ToUpper(GlobalLogging.Level(loggerName).String())
}

// 创建一个指定名称的日志记录器，无效的名称会导致panic
func MustGetLogger(loggerName string) *FabricLogger {
	return GlobalLogging.Logger(loggerName)
}

// 修改日志规格说明
func ActivateSpec(spec string) {
	err := GlobalLogging.ActivateSpec(spec)
	if err != nil {
		panic(err)
	}
}
