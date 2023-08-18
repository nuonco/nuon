package main

import (
	"github.com/powertoolsdev/mono/services/ctl-api/cmd"
)

// @title		Nuon API
// @version v1
// @description	This is a sample server celler server.
// @contact.name	Nuon Support
// @contact.email	support@nuon.co
// @BasePath	/api/v1
//
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description				Bearer token to access the api from

//go:generate -command swag go run github.com/swaggo/swag/cmd/swag
//go:generate swag init --parseGoList
//go:generate -command swagger go run github.com/go-swagger/go-swagger/cmd/swagger
//go:generate swagger validate ./docs/swagger.json
func main() {
	cmd.Execute()
}
