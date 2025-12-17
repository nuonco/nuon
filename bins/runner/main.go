package main

import (
	"github.com/nuonco/nuon/bins/runner/cmd"
)

//go:generate ./generate.sh
func main() {
	cmd.Execute()
}
