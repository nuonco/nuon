package executors

import (
	"fmt"
	"strings"

	"github.com/powertoolsdev/mono/pkg/generics"
)

type AzureSettings struct {
	Location                 string `json:"string"`
	SubscriptionID           string `json:"subscription_id"`
	SubscriptionTenantID     string `json:"subscription_tenant_id"`
	ServicePrincipalAppID    string `json:"service_principal_app_id"`
	ServicePrincipalPassword string `json:"service_principal_password"`
}

func (a AzureSettings) Validate() error {
	availableLocations := []string{
		"eastus",
		"eastus2",
		"southcentralus",
		"westus2",
		"westus3",
		"australiaeast",
		"southeastasia",
		"northeurope",
		"swedencentral",
		"uksouth",
		"westeurope",
		"centralus",
		"southafricanorth",
		"centralindia",
		"eastasia",
		"japaneast",
		"koreacentral",
		"canadacentral",
		"francecentral",
		"germanywestcentral",
		"italynorth",
		"norwayeast",
		"polandcentral",
		"switzerlandnorth",
		"uaenorth",
		"brazilsouth",
		"centraluseuap",
		"israelcentral",
		"qatarcentral",
		"centralusstage",
		"eastusstage",
		"eastus2stage",
		"northcentralusstage",
		"southcentralusstage",
		"westusstage",
		"westus2stage",
		"asia",
		"asiapacific",
		"australia",
		"brazil",
		"canada",
		"europe",
		"france",
		"germany",
		"global",
		"india",
		"israel",
		"italy",
		"japan",
		"korea",
		"newzealand",
		"norway",
		"poland",
		"qatar",
		"singapore",
		"southafrica",
		"sweden",
		"switzerland",
		"uae",
		"uk",
		"unitedstates",
		"unitedstateseuap",
		"eastasiastage",
		"southeastasiastage",
		"brazilus",
		"eastusstg",
		"northcentralus",
		"westus",
		"japanwest",
		"jioindiawest",
		"eastus2euap",
		"westcentralus",
		"southafricawest",
		"australiacentral",
		"australiacentral2",
		"australiasoutheast",
		"jioindiacentral",
		"koreasouth",
		"southindia",
		"westindia",
		"canadaeast",
		"francesouth",
		"germanynorth",
		"norwaywest",
		"switzerlandwest",
		"ukwest",
		"uaecentral",
		"brazilsoutheast",
	}

	if !generics.SliceContains(a.Location, availableLocations) {
		return fmt.Errorf("unsupported location %s, must be one of %s", a.Location, strings.Join(availableLocations, ", "))
	}

	return nil
}
