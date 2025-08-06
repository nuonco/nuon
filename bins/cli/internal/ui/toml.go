package ui

import (
	"bytes"
	"fmt"

	"github.com/pelletier/go-toml"
)

func PrintTOML(data interface{}) {
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	enc.SetTagName("mapstructure")

	_ = enc.Encode(data)

	fmt.Println(buf.String())
}
