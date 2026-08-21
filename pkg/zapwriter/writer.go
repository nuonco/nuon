package zapwriter

import (
	"bufio"
	"bytes"

	"github.com/pkg/errors"
	"go.uber.org/zap/zapcore"
)

func (z *zapWriter) Write(byts []byte) (int, error) {
	if z.lineBuffered {
		return z.writeBuffered(byts)
	}

	buf := bytes.NewBuffer(byts)

	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		z.logLine(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return 0, errors.Wrap(err, "unable to scan output")
	}

	return len(byts), nil
}

func (z *zapWriter) writeBuffered(byts []byte) (int, error) {
	z.mu.Lock()
	defer z.mu.Unlock()

	z.buffer = append(z.buffer, byts...)
	for {
		lineEnd := bytes.IndexByte(z.buffer, '\n')
		if lineEnd < 0 {
			break
		}

		line := z.buffer[:lineEnd]
		line = bytes.TrimSuffix(line, []byte{'\r'})
		z.logLine(string(line))
		z.buffer = z.buffer[lineEnd+1:]
	}

	return len(byts), nil
}

func (z *zapWriter) Flush() {
	z.mu.Lock()
	defer z.mu.Unlock()

	if len(z.buffer) == 0 {
		return
	}
	z.logLine(string(z.buffer))
	z.buffer = nil
}

func (z *zapWriter) logLine(inputLine string) {
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
