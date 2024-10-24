package zaphclog

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Assumed, we'll get key-value pairs as arguments.
// Code below prevents a panic, if wrong arguments set received.
func convertToZapAny(args ...interface{}) []zapcore.Field {
	fields := []zapcore.Field{}
	for i := len(args); i > 0; i -= 2 {
		left := i - 2
		if left < 0 {
			left = 0
		}

		items := args[left:i]

		switch l := len(items); l {
		case 2:
			k, ok := items[0].(string)
			if ok {
				fields = append(fields, zap.Any(k, items[1]))
			} else {
				fields = append(fields, zap.Any(fmt.Sprintf("arg%d", i-1), items[1]))
				fields = append(fields, zap.Any(fmt.Sprintf("arg%d", left), items[0]))
			}
		case 1:
			fields = append(fields, zap.Any(fmt.Sprintf("arg%d", left), items[0]))
		}
	}

	return fields
}
