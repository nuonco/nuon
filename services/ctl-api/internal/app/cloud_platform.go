package app

import "fmt"

type CloudPlatformRegion struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	DisplayName string `json:"display_name"`
	Icon        string `json:"icon"`
}

type CloudPlatform string

const (
	CloudPlatformAWS     CloudPlatform = "aws"
	CloudPlatformAzure   CloudPlatform = "azure"
	CloudPlatformUnknown CloudPlatform = "unknown"
)

func (c CloudPlatform) azureLocations() []CloudPlatformRegion {
	return []CloudPlatformRegion{
		{
			Name:        "East US",
			Value:       "eastus",
			DisplayName: "(US) East US",
			Icon:        "flag-US",
		},
		{
			Name:        "East US 2",
			Value:       "eastus2",
			DisplayName: "(US) East US 2",
			Icon:        "flag-US",
		},
		{
			Name:        "South Central US",
			Value:       "southcentralus",
			DisplayName: "(US) South Central US",
			Icon:        "flag-US",
		},
		{
			Name:        "West US 2",
			Value:       "westus2",
			DisplayName: "(US) West US 2",
			Icon:        "flag-US",
		},
		{
			Name:        "West US 3",
			Value:       "westus3",
			DisplayName: "(US) West US 3",
			Icon:        "flag-US",
		},
		{
			Name:        "Australia East",
			Value:       "australiaeast",
			DisplayName: "(Asia Pacific) Australia East",
			Icon:        "flag-AU",
		},
		{
			Name:        "Southeast Asia",
			Value:       "southeastasia",
			DisplayName: "(Asia Pacific) Southeast Asia",
			Icon:        "flag-SG",
		},
		{
			Name:        "North Europe",
			Value:       "northeurope",
			DisplayName: "(Europe) North Europe",
			Icon:        "flag-IE",
		},
		{
			Name:        "Sweden Central",
			Value:       "swedencentral",
			DisplayName: "(Europe) Sweden Central",
			Icon:        "flag-SE",
		},
		{
			Name:        "UK South",
			Value:       "uksouth",
			DisplayName: "(Europe) UK South",
			Icon:        "flag-GB",
		},
		{
			Name:        "West Europe",
			Value:       "westeurope",
			DisplayName: "(Europe) West Europe",
			Icon:        "flag-NL",
		},
		{
			Name:        "Central US",
			Value:       "centralus",
			DisplayName: "(US) Central US",
			Icon:        "flag-US",
		},
		{
			Name:        "South Africa North",
			Value:       "southafricanorth",
			DisplayName: "(Africa) South Africa North",
			Icon:        "flag-ZA",
		},
		{
			Name:        "Central India",
			Value:       "centralindia",
			DisplayName: "(Asia Pacific) Central India",
			Icon:        "flag-IN",
		},
		{
			Name:        "East Asia",
			Value:       "eastasia",
			DisplayName: "(Asia Pacific) East Asia",
			Icon:        "flag-HK",
		},
		{
			Name:        "Japan East",
			Value:       "japaneast",
			DisplayName: "(Asia Pacific) Japan East",
			Icon:        "flag-JP",
		},
		{
			Name:        "Korea Central",
			Value:       "koreacentral",
			DisplayName: "(Asia Pacific) Korea Central",
			Icon:        "flag-KR",
		},
		{
			Name:        "Canada Central",
			Value:       "canadacentral",
			DisplayName: "(Canada) Canada Central",
			Icon:        "flag-CA",
		},
		{
			Name:        "France Central",
			Value:       "francecentral",
			DisplayName: "(Europe) France Central",
			Icon:        "flag-FR",
		},
		{
			Name:        "Germany West Central",
			Value:       "germanywestcentral",
			DisplayName: "(Europe) Germany West Central",
			Icon:        "flag-DE",
		},
		{
			Name:        "Norway East",
			Value:       "norwayeast",
			DisplayName: "(Europe) Norway East",
			Icon:        "flag-NO",
		},
		{
			Name:        "Poland Central",
			Value:       "polandcentral",
			DisplayName: "(Europe) Poland Central",
			Icon:        "flag-PL",
		},
		{
			Name:        "Switzerland North",
			Value:       "switzerlandnorth",
			DisplayName: "(Europe) Switzerland North",
			Icon:        "flag-CH",
		},
		{
			Name:        "UAE North",
			Value:       "uaenorth",
			DisplayName: "(Middle East) UAE North",
			Icon:        "flag-AE",
		},
		{
			Name:        "Brazil South",
			Value:       "brazilsouth",
			DisplayName: "(South America) Brazil South",
			Icon:        "flag-BR",
		},
		{
			Name:        "Central US EUAP",
			Value:       "centraluseuap",
			DisplayName: "(US) Central US EUAP",
			Icon:        "flag-US",
		},
		{
			Name:        "Qatar Central",
			Value:       "qatarcentral",
			DisplayName: "(Middle East) Qatar Central",
			Icon:        "flag-QA",
		},
		{
			Name:        "Central US (Stage)",
			Value:       "centralusstage",
			DisplayName: "(US) Central US (Stage)",
			Icon:        "flag-US",
		},
		{
			Name:        "East US (Stage)",
			Value:       "eastusstage",
			DisplayName: "(US) East US (Stage)",
			Icon:        "flag-US",
		},
		{
			Name:        "East US 2 (Stage)",
			Value:       "eastus2stage",
			DisplayName: "(US) East US 2 (Stage)",
			Icon:        "flag-US",
		},
		{
			Name:        "North Central US (Stage)",
			Value:       "northcentralusstage",
			DisplayName: "(US) North Central US (Stage)",
			Icon:        "flag-US",
		},
		{
			Name:        "South Central US (Stage)",
			Value:       "southcentralusstage",
			DisplayName: "(US) South Central US (Stage)",
			Icon:        "flag-US",
		},
		{
			Name:        "West US (Stage)",
			Value:       "westusstage",
			DisplayName: "(US) West US (Stage)",
			Icon:        "flag-US",
		},
		{
			Name:        "West US 2 (Stage)",
			Value:       "westus2stage",
			DisplayName: "(US) West US 2 (Stage)",
			Icon:        "flag-US",
		},
		{
			Name:        "Asia",
			Value:       "asia",
			DisplayName: "Asia",
		},
		{
			Name:        "Asia Pacific",
			Value:       "asiapacific",
			DisplayName: "Asia Pacific",
			Icon:        "flag-SG",
		},
		{
			Name:        "Australia",
			Value:       "australia",
			DisplayName: "Australia",
			Icon:        "flag-AU",
		},
		{
			Name:        "Brazil",
			Value:       "brazil",
			DisplayName: "Brazil",
			Icon:        "flag-BR",
		},
		{
			Name:        "Canada",
			Value:       "canada",
			DisplayName: "Canada",
			Icon:        "flag-CA",
		},
		{
			Name:        "Europe",
			Value:       "europe",
			DisplayName: "Europe",
			Icon:        "flag-EU",
		},
		{
			Name:        "France",
			Value:       "france",
			DisplayName: "France",
			Icon:        "flag-FR",
		},
		{
			Name:        "Germany",
			Value:       "germany",
			DisplayName: "Germany",
			Icon:        "flag-DE",
		},
		{
			Name:        "Global",
			Value:       "global",
			DisplayName: "Global",
		},
		{
			Name:        "India",
			Value:       "india",
			DisplayName: "India",
			Icon:        "flag-IN",
		},
		{
			Name:        "Japan",
			Value:       "japan",
			DisplayName: "Japan",
			Icon:        "flag-JP",
		},
		{
			Name:        "Korea",
			Value:       "korea",
			DisplayName: "Korea",
			Icon:        "flag-KR",
		},
		{
			Name:        "Norway",
			Value:       "norway",
			DisplayName: "Norway",
			Icon:        "flag-NO",
		},
		{
			Name:        "Singapore",
			Value:       "singapore",
			DisplayName: "Singapore",
			Icon:        "flag-SG",
		},
		{
			Name:        "South Africa",
			Value:       "southafrica",
			DisplayName: "South Africa",
			Icon:        "flag-ZA",
		},
		{
			Name:        "Switzerland",
			Value:       "switzerland",
			DisplayName: "Switzerland",
			Icon:        "flag-CH",
		},
		{
			Name:        "United Arab Emirates",
			Value:       "uae",
			DisplayName: "United Arab Emirates",
			Icon:        "flag-AE",
		},
		{
			Name:        "United Kingdom",
			Value:       "uk",
			DisplayName: "United Kingdom",
			Icon:        "flag-GB",
		},
		{
			Name:        "United States",
			Value:       "unitedstates",
			DisplayName: "United States",
			Icon:        "flag-US",
		},
		{
			Name:        "United States EUAP",
			Value:       "unitedstateseuap",
			DisplayName: "United States EUAP",
			Icon:        "flag-US",
		},
		{
			Name:        "East Asia (Stage)",
			Value:       "eastasiastage",
			DisplayName: "(Asia Pacific) East Asia (Stage)",
			Icon:        "flag-HK",
		},
		{
			Name:        "Southeast Asia (Stage)",
			Value:       "southeastasiastage",
			DisplayName: "(Asia Pacific) Southeast Asia (Stage)",
			Icon:        "flag-SG",
		},
		{
			Name:        "Brazil US",
			Value:       "brazilus",
			DisplayName: "(South America) Brazil US",
			Icon:        "flag-BR",
		},
		{
			Name:        "East US STG",
			Value:       "eastusstg",
			DisplayName: "(US) East US STG",
			Icon:        "flag-US",
		},
		{
			Name:        "North Central US",
			Value:       "northcentralus",
			DisplayName: "(US) North Central US",
			Icon:        "flag-US",
		},
		{
			Name:        "West US",
			Value:       "westus",
			DisplayName: "(US) West US",
			Icon:        "flag-US",
		},
		{
			Name:        "Jio India West",
			Value:       "jioindiawest",
			DisplayName: "(Asia Pacific) Jio India West",
			Icon:        "flag-IN",
		},
		{
			Name:        "East US 2 EUAP",
			Value:       "eastus2euap",
			DisplayName: "(US) East US 2 EUAP",
			Icon:        "flag-US",
		},
		{
			Name:        "South Central US STG",
			Value:       "southcentralusstg",
			DisplayName: "(US) South Central US STG",
			Icon:        "flag-US",
		},
		{
			Name:        "West Central US",
			Value:       "westcentralus",
			DisplayName: "(US) West Central US",
			Icon:        "flag-US",
		},
		{
			Name:        "South Africa West",
			Value:       "southafricawest",
			DisplayName: "(Africa) South Africa West",
			Icon:        "flag-ZA",
		},
		{
			Name:        "Australia Central",
			Value:       "australiacentral",
			DisplayName: "(Asia Pacific) Australia Central",
			Icon:        "flag-AU",
		},
		{
			Name:        "Australia Central 2",
			Value:       "australiacentral2",
			DisplayName: "(Asia Pacific) Australia Central 2",
			Icon:        "flag-AU",
		},
		{
			Name:        "Australia Southeast",
			Value:       "australiasoutheast",
			DisplayName: "(Asia Pacific) Australia Southeast",
			Icon:        "flag-AU",
		},
		{
			Name:        "Japan West",
			Value:       "japanwest",
			DisplayName: "(Asia Pacific) Japan West",
			Icon:        "flag-JP",
		},
		{
			Name:        "Jio India Central",
			Value:       "jioindiacentral",
			DisplayName: "(Asia Pacific) Jio India Central",
			Icon:        "flag-IN",
		},
		{
			Name:        "Korea South",
			Value:       "koreasouth",
			DisplayName: "(Asia Pacific) Korea South",
			Icon:        "flag-KR",
		},
		{
			Name:        "South India",
			Value:       "southindia",
			DisplayName: "(Asia Pacific) South India",
			Icon:        "flag-IN",
		},
		{
			Name:        "West India",
			Value:       "westindia",
			DisplayName: "(Asia Pacific) West India",
			Icon:        "flag-IN",
		},
		{
			Name:        "Canada East",
			Value:       "canadaeast",
			DisplayName: "(Canada) Canada East",
			Icon:        "flag-CA",
		},
		{
			Name:        "France South",
			Value:       "francesouth",
			DisplayName: "(Europe) France South",
			Icon:        "flag-FR",
		},
		{
			Name:        "Germany North",
			Value:       "germanynorth",
			DisplayName: "(Europe) Germany North",
			Icon:        "flag-DE",
		},
		{
			Name:        "Norway West",
			Value:       "norwaywest",
			DisplayName: "(Europe) Norway West",
			Icon:        "flag-NO",
		},
		{
			Name:        "Switzerland West",
			Value:       "switzerlandwest",
			DisplayName: "(Europe) Switzerland West",
			Icon:        "flag-CH",
		},
		{
			Name:        "UK West",
			Value:       "ukwest",
			DisplayName: "(Europe) UK West",
			Icon:        "flag-GB",
		},
		{
			Name:        "UAE Central",
			Value:       "uaecentral",
			DisplayName: "(Middle East) UAE Central",
			Icon:        "flag-AE",
		},
		{
			Name:        "Brazil Southeast",
			Value:       "brazilsoutheast",
			DisplayName: "(South America) Brazil Southeast",
			Icon:        "flag-BR",
		},
		{
			DisplayName: "(Middle East) Israel Central",
			Icon:        "flag-IL",
			Name:        "Israel Central",
			Value:       "israelcentral",
		},
	}
}

func (c CloudPlatform) awsRegions() []CloudPlatformRegion {
	return []CloudPlatformRegion{
		{
			Name:        "us-east-1",
			Icon:        "flag-US",
			DisplayName: "US East (N. Virginia)",
			Value:       "us-east-1",
		},
		{
			Name:        "us-east-2",
			Icon:        "flag-US",
			DisplayName: "US East (Ohio)",
			Value:       "us-east-2",
		},
		{
			Name:        "us-west-1",
			Icon:        "flag-US",
			DisplayName: "US West (N. California)",
			Value:       "us-west-1",
		},
		{
			Name:        "us-west-2",
			Icon:        "flag-US",
			DisplayName: "US West (Oregon)",
			Value:       "us-west-2",
		},

		// africa
		{
			Name:        "af-south-1",
			Icon:        "flag-ZA",
			DisplayName: "Africa (Cape Town)",
			Value:       "af-south-1",
		},

		// asia
		{
			Name:        "ap-east-1",
			Icon:        "flag-HK",
			DisplayName: "Asia Pacific (Hong Kong)",
			Value:       "ap-east-1",
		},
		{
			Name:        "ap-south-2",
			Icon:        "flag-IN",
			DisplayName: "Asia Pacific (Hyderabad)",
			Value:       "ap-south-2",
		},
		{
			Name:        "ap-southeast-3",
			Icon:        "flag-ID",
			DisplayName: "Asia Pacific (Jakarta)",
			Value:       "ap-southeast-3",
		},
		{
			Name:        "ap-southeast-4",
			Icon:        "flag-AU",
			DisplayName: "Asia Pacific (Melbourne)",
			Value:       "ap-southeast-4",
		},
		{
			Name:        "ap-south-1",
			Icon:        "flag-IN",
			DisplayName: "Asia Pacific (Mumbai)",
			Value:       "ap-south-1",
		},
		{
			Name:        "ap-northeast-3",
			Icon:        "flag-JP",
			DisplayName: "Asia Pacific (Osaka)",
			Value:       "ap-northeast-3",
		},
		{
			Name:        "ap-northeast-2",
			Icon:        "flag-KR",
			DisplayName: "Asia Pacific (Seoul)",
			Value:       "ap-northeast-2",
		},
		{
			Name:        "ap-southeast-1",
			Icon:        "flag-SG",
			DisplayName: "Asia Pacific (Singapore)",
			Value:       "ap-southeast-1",
		},
		{
			Name:        "ap-southeast-2",
			Icon:        "flag-AU",
			DisplayName: "Asia Pacific (Sydney)",
			Value:       "ap-southeast-2",
		},
		{
			Name:        "ap-northeast-1",
			Icon:        "flag-JP",
			DisplayName: "Asia Pacific (Tokyo)",
			Value:       "ap-northeast-1",
		},

		// canada
		{
			Name:        "ca-central-1",
			Icon:        "flag-CA",
			DisplayName: "Canada (Central)",
			Value:       "ca-central-1",
		},
		{
			Name:        "ca-west-1",
			Icon:        "flag-CA",
			DisplayName: "Canada West (Calgery)",
			Value:       "ca-west-1",
		},

		// europe
		{
			Name:        "eu-central-1",
			Icon:        "flag-DE",
			DisplayName: "Europe (Frankfurt)",
			Value:       "eu-central-1",
		},
		{
			Name:        "eu-west-1",
			Icon:        "flag-IE",
			DisplayName: "Europe (Ireland)",
			Value:       "eu-west-1",
		},
		{
			Name:        "eu-west-2",
			Icon:        "flag-GB",
			DisplayName: "Europe (London)",
			Value:       "eu-west-2",
		},
		{
			Name:        "eu-south-1",
			Icon:        "flag-IT",
			DisplayName: "Europe (Milan)",
			Value:       "eu-south-1",
		},
		{
			Name:        "eu-west-3",
			Icon:        "flag-FR",
			DisplayName: "Europe (Paris)",
			Value:       "eu-west-3",
		},
		{
			Name:        "eu-south-2",
			Icon:        "flag-ES",
			DisplayName: "Europe (Spain)",
			Value:       "eu-south-2",
		},
		{
			Name:        "eu-north-1",
			Icon:        "flag-SE",
			DisplayName: "Europe (Stockholm)",
			Value:       "eu-north-1",
		},
		{
			Name:        "eu-central-2",
			Icon:        "flag-CH",
			DisplayName: "Europe (Zürich)",
			Value:       "eu-central-2",
		},

		// israel
		{
			Name:        "il-central-1",
			Icon:        "flag-IL",
			DisplayName: "Israel (Tel Aviv)",
			Value:       "il-central-1",
		},

		// middle east
		{
			Name:        "me-south-1",
			Icon:        "flag-BH",
			DisplayName: "Middle East (Bahrain)",
			Value:       "me-south-1",
		},
		{
			Name:        "me-central-1",
			Icon:        "flag-AE",
			DisplayName: "Middle East (UAE)",
			Value:       "me-central-1",
		},

		// south america
		{
			Name:        "sa-east-1",
			Icon:        "flag-BR",
			DisplayName: "South America (São Paulo)",
			Value:       "sa-east-1",
		},

		// gov cloud
		{
			Name:        "us-gov-east-1",
			Icon:        "flag-US",
			DisplayName: "AWS GovCloud (US-East)",
			Value:       "us-gov-east-1",
		},
		{
			Name:        "us-gov-west-1",
			Icon:        "flag-US",
			DisplayName: "AWS GovCloud (US-West)",
			Value:       "us-gov-west-1",
		},
	}
}

func (c CloudPlatform) Regions() []CloudPlatformRegion {
	switch c {
	case CloudPlatformAWS:
		return c.awsRegions()
	case CloudPlatformAzure:
		return c.azureLocations()
	default:
	}

	return []CloudPlatformRegion{}
}

func NewCloudPlatform(platform string) (CloudPlatform, error) {
	switch platform {
	case "aws":
		return CloudPlatformAWS, nil
	case "azure":
		return CloudPlatformAzure, nil
	default:
	}

	return CloudPlatformUnknown, fmt.Errorf("invalid cloud platform")
}
