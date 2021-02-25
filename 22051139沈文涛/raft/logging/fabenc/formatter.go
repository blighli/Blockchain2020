package fabenc

import (
	"fmt"
	"go.uber.org/zap/zapcore"
	"io"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
)

var formatRegexp = regexp.MustCompile(`%{(color|id|level|message|module|shortfunc|time)(?::(.*?))?}`)

type MultiFormatter struct {
	mutex      sync.RWMutex
	formatters []Formatter
}

func NewMultiFormatter(formatters ...Formatter) *MultiFormatter {
	return &MultiFormatter{
		formatters: formatters,
	}
}

func (m *MultiFormatter) Format(w io.Writer, entry zapcore.Entry, fields []zapcore.Field) {
	m.mutex.RLock()
	for i := range m.formatters {
		m.formatters[i].Format(w, entry, fields)
	}
	m.mutex.RUnlock()
}

func (m *MultiFormatter) SetFormatters(formatters []Formatter) {
	m.mutex.Lock()
	m.formatters = formatters
	m.mutex.Unlock()
}

type StringFormatter struct {
	Value string
}

func (s StringFormatter) Format(w io.Writer, entry zapcore.Entry, fields []zapcore.Field) {
	fmt.Fprintf(w, "%s", s.Value)
}

type MessageFormatter struct {
	FormatVerb string
}

func newMessageFormatter(s string) MessageFormatter {
	return MessageFormatter{FormatVerb: "%" + stringOrDefault(s, "s")}
}

func (m MessageFormatter) Format(w io.Writer, entry zapcore.Entry, fields []zapcore.Field) {
	fmt.Fprintf(w, m.FormatVerb, strings.TrimRight(entry.Message, "\n"))
}

type ColorFormatter struct {
	Bold  bool // 设置bold属性
	Reset bool // 重置颜色和属性
}

func newColorFormatter(s string) (ColorFormatter, error) {
	switch s {
	case "bold":
		return ColorFormatter{Bold: true}, nil
	case "reset":
		return ColorFormatter{Bold: false, Reset: true}, nil
	case "":
		return ColorFormatter{}, nil
	default:
		return ColorFormatter{}, fmt.Errorf("invalid color option: %s", s)
	}
}

// 根据日志级别返回对应的日志颜色
func (c ColorFormatter) LevelColor(l zapcore.Level) Color {
	switch l {
	case zapcore.DebugLevel:
		return ColorCyan
	case zapcore.InfoLevel:
		return ColorBlue
	case zapcore.WarnLevel:
		return ColorYellow
	case zapcore.ErrorLevel:
		return ColorRed
	case zapcore.DPanicLevel, zapcore.PanicLevel:
		return ColorMagenta
	case zapcore.FatalLevel:
		return ColorMagenta
	default:
		return ColorNone
	}
}

func (c ColorFormatter) Format(w io.Writer, entry zapcore.Entry, fields []zapcore.Field) {
	switch {
	case c.Reset:
		fmt.Fprintf(w, ResetColor())
	case c.Bold:
		fmt.Fprintf(w, c.LevelColor(entry.Level).Bold())
	default:
		fmt.Fprint(w, c.LevelColor(entry.Level).Normal())
	}
}

type TimeFormatter struct {
	Layout string
}

func newTimeFormatter(s string) TimeFormatter {
	return TimeFormatter{
		Layout: stringOrDefault(s, "2021-01-01T00:00:00.999Z07:00"),
	}
}

func (t TimeFormatter) Format(w io.Writer, entry zapcore.Entry, fields []zapcore.Field) {
	fmt.Fprintf(w, entry.Time.Format(t.Layout))
}

// 格式化zap日志记录器名
type ModuleFormatter struct {
	FormatVerb string
}

func newModuleFormatter(s string) ModuleFormatter {
	return ModuleFormatter{FormatVerb: "%" + stringOrDefault(s, "s")}
}

func (m ModuleFormatter) Format(w io.Writer, entry zapcore.Entry, fields []zapcore.Field) {
	fmt.Fprintf(w, m.FormatVerb, entry.LoggerName)
}

// 格式化创建日志的函数名称
type ShortFuncFormatter struct {
	FormatVerb string
}

func newShortFuncFormatter(s string) ShortFuncFormatter {
	return ShortFuncFormatter{FormatVerb: "%" + stringOrDefault(s, "s")}
}

func (s ShortFuncFormatter) Format(w io.Writer, entry zapcore.Entry, fields []zapcore.Field) {
	f := runtime.FuncForPC(entry.Caller.PC)
	if f == nil {
		fmt.Fprintf(w, s.FormatVerb, "(unknown)")
		return
	}

	fname := f.Name()
	funcIdx := strings.LastIndex(fname, ".")
	fmt.Fprintf(w, s.FormatVerb, fname[funcIdx+1:])
}

// 格式化日志级别
type LevelFormatter struct {
	FormatVerb string
}

func newLevelFormatter(s string) LevelFormatter {
	return LevelFormatter{FormatVerb: "%" + stringOrDefault(s, "s")}
}

func (l LevelFormatter) Format(w io.Writer, entry zapcore.Entry, fields []zapcore.Field) {
	fmt.Fprintf(w, l.FormatVerb, entry.Level.CapitalString())
}

var sequence uint64

// 格式化全局日志序列号
type SequenceFormatter struct {
	FormatVerb string
}

func newSequenceFormatter(s string) SequenceFormatter {
	return SequenceFormatter{FormatVerb: "%" + stringOrDefault(s, "d")}
}

func (s SequenceFormatter) Format(w io.Writer, entry zapcore.Entry, fields []zapcore.Field) {
	fmt.Fprintf(w, s.FormatVerb, atomic.AddUint64(&sequence, 1))
}

// 解析一个日志规格说明并返回一个formatter切片
//
// 支持的说明符如下：
//   - %{color} - level specific SGR color escape or SGR reset
//   - %{id} - a unique log sequence number
//   - %{level} - the log level of the entry
//   - %{message} - the log message
//   - %{module} - the zap logger name
//   - %{shortfunc} - the name of the function creating the log record
//   - %{time} - the time the log entry was created
//
// 说明符可以包括可选格式动词：
//   - color: reset|bold
//   - id: a fmt style numeric formatter without the leading %
//   - level: a fmt style string formatter without the leading %
//   - message: a fmt style string formatter without the leading %
//   - module: a fmt style string formatter without the leading %
//
func ParseFormat(spec string) ([]Formatter, error) {
	cursor := 0
	formatters := []Formatter{}
	matches := formatRegexp.FindAllStringSubmatchIndex(spec, -1)
	for _, m := range matches {
		start, end := m[0], m[1]
		verbStart, verbEnd := m[2], m[3]
		formatStart, formatEnd := m[4], m[5]

		if start > cursor {
			formatters = append(formatters, StringFormatter{Value: spec[cursor:start]})
		}

		var format string
		if formatStart >= 0 {
			format = spec[formatStart:formatEnd]
		}
		formatter, err := NewFormatter(spec[verbStart:verbEnd], format)
		if err != nil {
			return nil, err
		}
		formatters = append(formatters, formatter)
		cursor = end
	}
	if cursor != len(spec) {
		formatters = append(formatters, StringFormatter{Value: spec[cursor:]})
	}
	return formatters, nil
}

func NewFormatter(verb, format string) (Formatter, error) {
	switch verb {
	case "color":
		return newColorFormatter(format)
	case "id":
		return newSequenceFormatter(format), nil
	case "level":
		return newLevelFormatter(format), nil
	case "message":
		return newMessageFormatter(format), nil
	case "module":
		return newModuleFormatter(format), nil
	case "shortfunc":
		return newShortFuncFormatter(format), nil
	case "time":
		return newTimeFormatter(format), nil
	default:
		return nil, fmt.Errorf("unknown verb: %s", verb)
	}
}

func stringOrDefault(str, dflt string) string {
	if str != "" {
		return str
	}
	return dflt
}
