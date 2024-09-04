package main

import (
	"github.com/powertoolsdev/mono/bins/runner/cmd"
)

//go:generate ./generate.sh
func main() {
	cmd.Execute()
}
