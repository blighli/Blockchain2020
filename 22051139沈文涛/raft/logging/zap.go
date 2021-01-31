package logging

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zapgrpc"
)

type FabricLogger struct {
	s *zap.SugaredLogger
}

func (f *FabricLogger) DPanic(args ...interface{})                    { f.s.DPanicf(formatArgs(args)) }
func (f *FabricLogger) DPanicf(template string, args ...interface{})  { f.s.DPanicf(template, args...) }
func (f *FabricLogger) DPanicw(msg string, kvPairs ...interface{})    { f.s.DPanicw(msg, kvPairs...) }
func (f *FabricLogger) Debug(args ...interface{})                     { f.s.Debugf(formatArgs(args)) }
func (f *FabricLogger) Debugf(template string, args ...interface{})   { f.s.Debugf(template, args...) }
func (f *FabricLogger) Debugw(msg string, kvPairs ...interface{})     { f.s.Debugw(msg, kvPairs...) }
func (f *FabricLogger) Error(args ...interface{})                     { f.s.Errorf(formatArgs(args)) }
func (f *FabricLogger) Errorf(template string, args ...interface{})   { f.s.Errorf(template, args...) }
func (f *FabricLogger) Errorw(msg string, kvPairs ...interface{})     { f.s.Errorw(msg, kvPairs...) }
func (f *FabricLogger) Fatal(args ...interface{})                     { f.s.Fatalf(formatArgs(args)) }
func (f *FabricLogger) Fatalf(template string, args ...interface{})   { f.s.Fatalf(template, args...) }
func (f *FabricLogger) Fatalw(msg string, kvPairs ...interface{})     { f.s.Fatalw(msg, kvPairs...) }
func (f *FabricLogger) Info(args ...interface{})                      { f.s.Infof(formatArgs(args)) }
func (f *FabricLogger) Infof(template string, args ...interface{})    { f.s.Infof(template, args...) }
func (f *FabricLogger) Infow(msg string, kvPairs ...interface{})      { f.s.Infow(msg, kvPairs...) }
func (f *FabricLogger) Panic(args ...interface{})                     { f.s.Panic(formatArgs(args)) }
func (f *FabricLogger) Panicf(template string, args ...interface{})   { f.s.Panicf(template, args...) }
func (f *FabricLogger) Panicw(msg string, kvPairs ...interface{})     { f.s.Panicw(msg, kvPairs...) }
func (f *FabricLogger) Warn(args ...interface{})                      { f.s.Warnf(formatArgs(args)) }
func (f *FabricLogger) Warnf(template string, args ...interface{})    { f.s.Warnf(template, args...) }
func (f *FabricLogger) Warnw(msg string, kvPairs ...interface{})      { f.s.Warnw(msg, kvPairs...) }
func (f *FabricLogger) Warning(args ...interface{})                   { f.s.Warnf(formatArgs(args)) }
func (f *FabricLogger) Warningf(template string, args ...interface{}) { f.s.Warnf(template, args...) }

func NewZapLogger(core zapcore.Core, options ...zap.Option) *zap.Logger {
	return zap.New(
		core,
		append([]zap.Option{
			zap.AddCaller(),
			zap.AddStacktrace(zapcore.ErrorLevel),
		}, options...)...,
	)
}

// 创建一个 grpc.Logger 用于代理一个 zap.Logger
func NewGRPCLogger(l *zap.Logger) *zapgrpc.Logger {
	l = l.WithOptions(
		zap.AddCaller(),
		zap.AddCallerSkip(3),
	)
	return zapgrpc.NewLogger(l, zapgrpc.WithDebug())
}

func NewFabricLogger(l *zap.Logger, options ...zap.Option) *FabricLogger {
	return &FabricLogger{
		s: l.WithOptions(append(options, zap.AddCallerSkip(1))...).Sugar(),
	}
}

func formatArgs(args []interface{}) string {
	return strings.TrimSuffix(fmt.Sprintln(args...), "\n")
}

// 向后兼容
func (f *FabricLogger) Critical(args ...interface{})                   { f.s.Errorf(formatArgs(args)) }
func (f *FabricLogger) Criticalf(template string, args ...interface{}) { f.s.Errorf(template, args...) }
func (f *FabricLogger) Notice(args ...interface{})                     { f.s.Infof(formatArgs(args)) }
func (f *FabricLogger) Noticef(template string, args ...interface{})   { f.s.Infof(template, args...) }

func (f *FabricLogger) Zap() *zap.Logger { return f.s.Desugar() }

func (f *FabricLogger) IsEnabledFor(level zapcore.Level) bool {
	return f.s.Desugar().Core().Enabled(level)
}

func (f *FabricLogger) With(args ...interface{}) *FabricLogger {
	return &FabricLogger{s: f.s.With(args...)}
}

func (f *FabricLogger) WithOptions(opts ...zap.Option) *FabricLogger {
	l := f.s.Desugar().WithOptions(opts...)
	return &FabricLogger{s: l.Sugar()}
}
