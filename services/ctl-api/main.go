package main

import "github.com/powertoolsdev/mono/services/ctl-api/cmd"

//go:generate ./generate.sh
func main() {
	cmd.Execute()
}
