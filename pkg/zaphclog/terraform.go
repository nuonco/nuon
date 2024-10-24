package zaphclog

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type terraformLineOutput struct {
	Level string `json:"@level"`
	Msg   string `json:"@message"`
}

func (z *zaphclogWriter) writeTerraform(byts []byte) error {
	var tfLine terraformLineOutput
	if err := json.Unmarshal(byts, &tfLine); err != nil {
		return errors.Wrap(err, "assuming not terraform json")
	}

	switch tfLine.Level {
	case "info":
		z.l.Info(tfLine.Msg)
	case "error":
		z.l.Error(tfLine.Msg)
	case "warning", "warn":
		z.l.Warn(tfLine.Msg)
	}

	return nil
}
