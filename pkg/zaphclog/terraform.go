package zaphclog

import (
	"encoding/json"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type terraformLineOutput struct {
	Level     string `mapstructure:"@level"`
	Msg       string `mapstructure:"@message"`
	Timestamp string `mapsturcture:"@timestamp"`

	Other map[string]interface{} `mapstructure:",remain"`
}

func (z *zaphclogWriter) writeTerraform(byts []byte) error {
	var obj map[string]interface{}
	if err := json.Unmarshal(byts, &obj); err != nil {
		return errors.Wrap(err, "unable to parse to attrs")
	}

	var tfLine terraformLineOutput
	if err := mapstructure.Decode(obj, &tfLine); err != nil {
		return errors.Wrap(err, "unable to decode to terraform output")
	}

	attrs := make([]zapcore.Field, 0, len(tfLine.Other))
	for k, v := range tfLine.Other {
		attrs = append(attrs, zap.Any(k, v))
	}

	switch tfLine.Level {
	case "info":
		z.zl.Info(tfLine.Msg, attrs...)
	case "error":
		z.zl.Error(tfLine.Msg, attrs...)
	case "warning", "warn":
		z.zl.Warn(tfLine.Msg, attrs...)
	}

	return nil
}
