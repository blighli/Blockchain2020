package fabenc

import (
	zaplogfmt "github.com/sykesm/zap-logfmt"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"io"
	"time"
)

// A FormatEncoder is a zapcore.Encoder that formats log records according to a go-logging based format specifier
type FormatEncoder struct {
	zapcore.Encoder
	formatters []Formatter
	pool       buffer.Pool
}

type Formatter interface {
	Format(w io.Writer, entry zapcore.Entry, fields []zapcore.Field)
}

// 克隆一个FormatEncoder实例
func (f *FormatEncoder) Clone() zapcore.Encoder {
	return &FormatEncoder{
		Encoder:    f.Encoder.Clone(),
		formatters: f.formatters,
		pool:       f.pool,
	}
}

func (f *FormatEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	buf := f.pool.Get()
	for _, f := range f.formatters {
		f.Format(buf, entry, fields)
	}

	encoderFields, err := f.Encoder.EncodeEntry(entry, fields)
	if err != nil {
		return nil, err
	}
	if buf.Len() > 0 && encoderFields.Len() != 1 {
		buf.AppendString(" ")
	}
	buf.AppendString(encoderFields.String())
	encoderFields.Free()

	return buf, nil
}

func NewFormatEncoder(formatters ...Formatter) *FormatEncoder {
	return &FormatEncoder{
		Encoder: zaplogfmt.NewEncoder(zapcore.EncoderConfig{
			MessageKey:     "", // disable
			LevelKey:       "", // disable
			TimeKey:        "", // disable
			NameKey:        "", // disable
			CallerKey:      "", // disable
			StacktraceKey:  "", // disable
			LineEnding:     "\n",
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Format("2006-01-02T15:04:05.999Z07:00"))
			},
		}),
		formatters: formatters,
		pool:       buffer.NewPool(),
	}
}
