package credentials

import (
	"log"

	azlog "github.com/Azure/azure-sdk-for-go/sdk/azcore/log"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

func init() {
	// print log output to stdout
	azlog.SetListener(func(event azlog.Event, s string) {
		log.Println(s)
	})
	// include only azidentity credential logs
	azlog.SetEvents(azidentity.EventAuthentication)
}
