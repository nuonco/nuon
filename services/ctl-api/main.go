package main

import "github.com/powertoolsdev/mono/services/ctl-api/cmd"

//go:generate -command gen go run github.com/powertoolsdev/mono/services/ctl-api/cmd/gen
//go:generate gen
func main() {
	cmd.Execute()
}
