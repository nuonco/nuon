package json

import (
	"bytes"
	"context"
	jsn "encoding/json"
	"fmt"
	"io"

	"github.com/powertoolsdev/mono/pkg/terraform/archive"
)

func (d *json) Unpack(ctx context.Context, cb archive.Callback) error {
	// prettify the json, to make debugging easier
	var obj map[string]interface{}
	if err := jsn.Unmarshal(d.Byts, &obj); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}

	byts, err := jsn.MarshalIndent(obj, "", "    ")
	if err != nil {
		return fmt.Errorf("unable to create json: %w", err)
	}

	buf := bytes.NewBuffer(byts)
	rc := io.NopCloser(buf)

	if err := cb(ctx, d.FileName, rc); err != nil {
		return fmt.Errorf("unable to pass callback for json: %w", err)
	}
	return nil
}
