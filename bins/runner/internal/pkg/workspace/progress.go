package workspace

import "go.uber.org/zap"

type progressWriter struct {
	l *zap.Logger
}

func (p *progressWriter) Write(byts []byte) (int, error) {
	p.l.Info(string(byts))

	return len(byts), nil
}
