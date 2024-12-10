package zapwriter

import (
	"bufio"
	"bytes"
	"io"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ io.Writer = (*zapWriter)(nil)

func New(l *zap.Logger, level zapcore.Level, prefix string) *zapWriter {
	return &zapWriter{
		l:      l,
		level:  level,
		prefix: prefix,
	}
}

type zapWriter struct {
	l      *zap.Logger
	level  zapcore.Level
	prefix string
}

func (z *zapWriter) Write(byts []byte) (int, error) {
	buf := bytes.NewBuffer(byts)

	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		msg := z.prefix + scanner.Text()

		switch z.level {
		case zapcore.ErrorLevel:
			z.l.Error(msg)
		case zapcore.InfoLevel:
			z.l.Info(msg)
		case zapcore.DebugLevel:
			z.l.Debug(msg)
		default:
			z.l.Info(msg)
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, errors.Wrap(err, "unable to scan output")
	}

	return len(byts), nil
}
