package helm

/*
Copyright The Helm Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"context"
	"log/slog"
	"os"
)

type DebugEnabledFunc func() bool

type DebugCheckHandler struct {
	handler      slog.Handler
	debugEnabled DebugEnabledFunc
}

func (h *DebugCheckHandler) Enabled(_ context.Context, level slog.Level) bool {
	if level == slog.LevelDebug {
		return h.debugEnabled()
	}
	return true
}

func (h *DebugCheckHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.handler.Handle(ctx, r)
}

func (h *DebugCheckHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &DebugCheckHandler{
		handler:      h.handler.WithAttrs(attrs),
		debugEnabled: h.debugEnabled,
	}
}

func (h *DebugCheckHandler) WithGroup(name string) slog.Handler {
	return &DebugCheckHandler{
		handler:      h.handler.WithGroup(name),
		debugEnabled: h.debugEnabled,
	}
}

func NewLogger(debugEnabled DebugEnabledFunc) *slog.Logger {
	baseHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})

	dynamicHandler := &DebugCheckHandler{
		handler:      baseHandler,
		debugEnabled: debugEnabled,
	}

	return slog.New(dynamicHandler)
}
