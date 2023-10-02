package main

import (
	"github.com/powertoolsdev/mono/services/ctl-api/cmd"
)

//	@title			Nuon API
//	@version		v1
//	@description	API for managing nuon apps and installs.
//	@contact.name	Nuon Support
//	@contact.email	support@nuon.co
//	@BasePath		/
//	@schemes		https

//go:generate ./generate.sh
func main() {
	cmd.Execute()
}
