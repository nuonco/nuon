package credentials

import (
	"log"

	azlog "github.com/Azure/azure-sdk-for-go/sdk/azcore/log"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

func init() {
	azlog.SetListener(func(event azlog.Event, s string) {
		log.Println(s)
	})
	azlog.SetEvents(azidentity.EventAuthentication)
}
