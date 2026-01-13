package ui

import (
	"bytes"
	"fmt"

	"github.com/pelletier/go-toml/v2"
)

func PrintTOML(data interface{}) {
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)

	_ = enc.Encode(data)

	fmt.Println(buf.String())
}
