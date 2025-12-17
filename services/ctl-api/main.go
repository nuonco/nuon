package main

import "github.com/nuonco/nuon/services/ctl-api/cmd"

//go:generate -command gen go run github.com/nuonco/nuon/services/ctl-api/cmd/gen
//go:generate gen
func main() {
	cmd.Execute()
}
