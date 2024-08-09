package main

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/grafana/codejen"
)

// SlashHeaderMapper produces a FileMapper that injects a comment header onto
// a [codejen.File] indicating the main generator that produced it (via the provided
// maingen, which should be a path) and the jenny or jennies that constructed the
// file.
func SlashHeaderMapper(maingen string) codejen.FileMapper {
	return func(f codejen.File) (codejen.File, error) {
		buf := new(bytes.Buffer)
		if err := tmpls.Lookup("gen_header.tmpl").Execute(buf, tvars_gen_header{
			MainGenerator: filepath.ToSlash(maingen),
			Using:         f.From,
			Leader:        "// ",
		}); err != nil {
			return codejen.File{}, fmt.Errorf("failed executing gen header template: %w", err)
		}
		fmt.Fprint(buf, string(f.Data))
		f.Data = buf.Bytes()
		return f, nil
	}
}
