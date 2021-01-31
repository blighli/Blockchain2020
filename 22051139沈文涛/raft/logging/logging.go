package logging

import (
	"fmt"
	zaplogfmt "github.com/sykesm/zap-logfmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"raft/logging/fabenc"
	"regexp"
	"sync"
)

const (
	CONSOLE = iota
	JSON
	LOGFMT
)

// 定义有效的日志名称，[:alnum:]匹配数字、大小写字母
var loggerNameRegexp = regexp.MustCompile(`^[[:alnum:]_#:-]+(\.[[:alnum:]_#:-]+)*$`)

// 为 Logging 实例提供依赖
type Config struct {
	// 日志记录的格式，不提供则使用默认格式
	Format string
	// 日志记录级别，默认为INFO
	LogSpec string
	// 用于编码和格式化日志记录的接收器，默认为 os.Stderr
	Writer io.Writer
}

type Logging struct {
	*LoggerLevels
	mutex          sync.RWMutex
	encoding       Encoding
	encoderConfig  zapcore.EncoderConfig
	multiFormatter *fabenc.MultiFormatter
	writer         zapcore.WriteSyncer
	observer       Observer
}

// 创建一个新的 Logging 并根据配置进行初始化
func New(c Config) (*Logging, error) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.NameKey = "name"
	l := &Logging{
		LoggerLevels: &LoggerLevels{
			defaultLevel: defaultLevel,
		},
		encoderConfig:  encoderConfig,
		multiFormatter: fabenc.NewMultiFormatter(),
	}
	// 应用配置
	err := l.Apply(c)
	if err != nil {
		return nil, err
	}
	return l, nil
}

// 将配置应用于日志系统
func (l *Logging) Apply(c Config) error {
	err := l.SetFormat(c.Format)
	if err != nil {
		return err
	}

	if c.LogSpec == "" {
		c.LogSpec = os.Getenv("FABRIC_LOGGING_SPEC")
	}
	if c.LogSpec == "" {
		c.LogSpec = defaultLevel.String()
	}

	err = l.LoggerLevels.ActivateSpec(c.LogSpec)
	if err != nil {
		return nil
	}

	if c.Writer == nil {
		c.Writer = os.Stderr
	}
	l.SetWriter(c.Writer)

	return nil
}

// 设置日志记录格式和编码方式
func (l *Logging) SetFormat(format string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if format == "" {
		format = defaultFormat
	}
	if format == "json" {
		l.encoding = JSON
		return nil
	}
	if format == "logfmt" {
		l.encoding = LOGFMT
		return nil
	}

	formatters, err := fabenc.ParseFormat(format)
	if err != nil {
		return err
	}
	l.multiFormatter.SetFormatters(formatters)
	l.encoding = CONSOLE
	return nil
}

func (l *Logging) SetWriter(w io.Writer) {
	var writer zapcore.WriteSyncer
	switch t := w.(type) {
	case *os.File:
		writer = zapcore.Lock(t)
	case zapcore.WriteSyncer:
		writer = t
	default:
		writer = zapcore.AddSync(w)
	}
	l.mutex.Lock()
	l.writer = writer
	l.mutex.Unlock()
}

// 实例化一个指定名称的 zap.Logger
func (l *Logging) ZapLogger(name string) *zap.Logger {
	if !isValidLoggerName(name) {
		panic(fmt.Sprintf("invalid logger name: %s", name))
	}

	l.mutex.RLock()
	core := &Core{
		LevelEnabler: l.LoggerLevels,
		Levels:       l.LoggerLevels,
		Encoders: map[Encoding]zapcore.Encoder{
			CONSOLE: fabenc.NewFormatEncoder(l.multiFormatter),
			JSON:    zapcore.NewJSONEncoder(l.encoderConfig),
			LOGFMT:  zaplogfmt.NewEncoder(l.encoderConfig),
		},
		Selector: l,
		Output:   l,
		Observer: l,
	}
	l.mutex.RUnlock()

	return NewZapLogger(core).Named(name)
}

func (l *Logging) Sync() error {
	l.mutex.RLock()
	w := l.writer
	l.mutex.RUnlock()
	return w.Sync()
}

func (l *Logging) Write(b []byte) (int, error) {
	l.mutex.RLock()
	w := l.writer
	l.mutex.RUnlock()
	return w.Write(b)
}

func (l *Logging) Encoding() Encoding {
	l.mutex.RLock()
	e := l.encoding
	l.mutex.RUnlock()
	return e
}

func (l *Logging) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) {
	l.mutex.RLock()
	observer := l.observer
	l.mutex.RUnlock()

	if observer != nil {
		observer.Check(e, ce)
	}
}

func (l *Logging) WriteEntry(e zapcore.Entry, fields []zapcore.Field) {
	l.mutex.RLock()
	observer := l.observer
	l.mutex.RUnlock()
	if observer != nil {
		observer.WriteEntry(e, fields)
	}
}

func (l *Logging) SetObserver(observer Observer) {
	l.mutex.RLock()
	l.observer = observer
	l.mutex.RUnlock()
}

func (l *Logging) Logger(name string) *FabricLogger {
	zapLogger := l.ZapLogger(name)

	fabricLogger := &FabricLogger{
		s: zapLogger.WithOptions(append([]zap.Option{}, zap.AddCallerSkip(1))...).Sugar(),
	}
	return fabricLogger
}
