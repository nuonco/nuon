package temporalzap

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var tests = map[string]struct {
	msg         string
	vals        []interface{}
	errExpected error
}{
	"no vals": {
		msg:  "something",
		vals: []interface{}{},
	},
	"odd vals": {
		msg:         "something",
		vals:        []interface{}{"odd"},
		errExpected: errors.New("odd number of keyvals pairs"),
	},
	"non-string key": {
		msg:  "something",
		vals: []interface{}{struct{ name string }{name: "test"}, "something"},
	},
	"with vals": {
		msg:  "something",
		vals: []interface{}{"test", float64(1)},
	},
	"with many vals": {
		msg: "something",
		vals: []interface{}{
			"test_num", float64(1),
			"test_string", "string",
			"test_struct", map[string]interface{}{}, // bc json
		},
	},
}

func TestDebug(t *testing.T) {
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var out map[string]interface{}
			buf := &bytes.Buffer{}
			l := NewLogger(testableZapLogger(buf))

			l.Debug(test.msg, test.vals...)
			err := json.Unmarshal(buf.Bytes(), &out)
			assert.NoError(t, err)
			assert.Equal(t, test.msg, out["msg"])
			assert.Equal(t, "debug", out["level"])

			if test.errExpected != nil {
				assert.Contains(t, out["error"], test.errExpected.Error())
				return
			}

			testKeyvals(t, test.vals, out)
		})
	}
}

func TestInfo(t *testing.T) {
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var out map[string]interface{}
			buf := &bytes.Buffer{}
			l := NewLogger(testableZapLogger(buf))

			l.Info(test.msg, test.vals...)
			err := json.Unmarshal(buf.Bytes(), &out)
			assert.NoError(t, err)
			assert.Equal(t, test.msg, out["msg"])
			assert.Equal(t, "info", out["level"])

			if test.errExpected != nil {
				assert.Contains(t, out["error"], test.errExpected.Error())
				return
			}

			testKeyvals(t, test.vals, out)
		})
	}
}

func TestWarn(t *testing.T) {
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var out map[string]interface{}
			buf := &bytes.Buffer{}
			l := NewLogger(testableZapLogger(buf))

			l.Warn(test.msg, test.vals...)
			err := json.Unmarshal(buf.Bytes(), &out)
			assert.NoError(t, err)
			assert.Equal(t, test.msg, out["msg"])
			assert.Equal(t, "warn", out["level"])

			if test.errExpected != nil {
				assert.Contains(t, out["error"], test.errExpected.Error())
				return
			}

			testKeyvals(t, test.vals, out)
		})
	}
}

func TestError(t *testing.T) {
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var out map[string]interface{}
			buf := &bytes.Buffer{}
			l := NewLogger(testableZapLogger(buf))

			l.Error(test.msg, test.vals...)
			err := json.Unmarshal(buf.Bytes(), &out)
			assert.NoError(t, err)
			assert.Equal(t, test.msg, out["msg"])
			assert.Equal(t, "error", out["level"])

			if test.errExpected != nil {
				assert.Contains(t, out["error"], test.errExpected.Error())
				return
			}

			testKeyvals(t, test.vals, out)
		})
	}
}

func TestWith(t *testing.T) {
	var out map[string]interface{}
	buf := &bytes.Buffer{}
	l := NewLogger(testableZapLogger(buf))

	l = l.With("a", "b", "c", "d")
	l.Debug(t.Name())
	err := json.Unmarshal(buf.Bytes(), &out)
	assert.NoError(t, err)
	assert.Equal(t, t.Name(), out["msg"])

	assert.Equal(t, "b", out["a"])
	assert.Equal(t, "d", out["c"])
}

func TestZapLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	l := testableZapLogger(buf)

	log := NewLogger(l)
	expectedLog := l.WithOptions(zap.AddCallerSkip(1))
	assert.Equal(t, expectedLog, log.ZapLogger())
}

func testableZapLogger(w io.Writer) *zap.Logger {
	ws := zapcore.AddSync(w)
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		ws,
		zapcore.DebugLevel,
	)
	return zap.New(core)
}

func testKeyvals(t *testing.T, vals []interface{}, out map[string]interface{}) {
	iters := len(vals) / 2
	for i := 0; i < iters; i += 1 {
		ix := i * 2
		key, ok := vals[ix].(string)
		if !ok {
			key = fmt.Sprintf("%v", vals[ix])
		}
		value := vals[ix+1]
		assert.Equal(t, value, out[key])
	}
}
