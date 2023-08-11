package main

import (
	"github.com/powertoolsdev/mono/services/ctl-api/cmd"
)

//go:generate -command swag go run github.com/swaggo/swag/cmd/swag
//go:generate swag init
func main() {
	cmd.Execute()
}
