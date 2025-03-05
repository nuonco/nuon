package zapwriter

import (
	"bufio"
	"bytes"

	"github.com/pkg/errors"
	"go.uber.org/zap/zapcore"
)

func (z *zapWriter) Write(byts []byte) (int, error) {
	buf := bytes.NewBuffer(byts)

	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		inputLine := scanner.Text()
		msg := inputLine
		if z.lineFormatter != nil {
			msg = z.lineFormatter(msg)
		}

		level := z.level
		if z.lineLeveler != nil {
			level = z.lineLeveler(inputLine)
		}

		switch level {
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
