package wrt

import (
	"bufio"
	"bytes"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logger struct {
	name     string
	zl       *zap.Logger
	isStderr bool
}

func (l *logger) Write(byts []byte) (int, error) {
	buf := bytes.NewBuffer(byts)

	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		attrs := []zapcore.Field{
			zap.String("wasm-module", l.name),
		}

		txt := scanner.Text()

		if l.isStderr {
			l.zl.Error(txt, attrs...)
		} else {
			l.zl.Info(txt, attrs...)
		}
	}

	return len(byts), nil
}
