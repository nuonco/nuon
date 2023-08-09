package ui

import (
	"context"
	"fmt"
	"os"

	"github.com/kyokomi/emoji"
	"go.uber.org/zap"
)

func Line(ctx context.Context, msg string, args ...interface{}) {
	log, err := FromContext(ctx)
	if err != nil {
		return
	}

	log.Step(msg, args...)
}

// Step records a step that happened, with a "check"
func Step(ctx context.Context, msg string, args ...interface{}) {
	log, err := FromContext(ctx)
	if err != nil {
		return
	}

	log.Step(msg, args...)
}

func (l *logger) Step(msg string, args ...interface{}) {
	if l.JSON {
		l.Zap.Info(fmt.Sprintf(msg, args...))
		return
	}

	emoji.Fprintf(os.Stderr, ":check_mark:"+msg+"\n", args...)
}

// Error records an error that happened, with a "x"
func Error(ctx context.Context, userErr error) {
	log, err := FromContext(ctx)
	if err != nil {
		return
	}

	log.Error(userErr)
}

func (l *logger) Error(err error) {
	if l.JSON {
		l.Zap.Error("error", zap.Error(err))
		return
	}

	emoji.Fprintf(os.Stderr, ":siren: %s\n", err.Error())
}
