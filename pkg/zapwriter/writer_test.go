package zapwriter

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestLineBufferedWriterPreservesLinesAcrossWrites(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	writer := NewWithOpts(zap.New(core), WithLineBuffering())

	_, err := writer.Write([]byte("first line\npartial "))
	require.NoError(t, err)
	require.Equal(t, []string{"first line"}, observedMessages(logs))

	_, err = writer.Write([]byte("line\r\nlast line"))
	require.NoError(t, err)
	require.Equal(t, []string{"first line", "partial line"}, observedMessages(logs))

	writer.Flush()
	require.Equal(t, []string{"first line", "partial line", "last line"}, observedMessages(logs))
}

func observedMessages(logs *observer.ObservedLogs) []string {
	entries := logs.All()
	messages := make([]string, len(entries))
	for idx, entry := range entries {
		messages[idx] = entry.Message
	}
	return messages
}
