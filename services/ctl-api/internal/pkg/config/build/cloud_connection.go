package build

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

var (
	ecrAccountPattern  = regexp.MustCompile(`^([0-9]{12})\.dkr\.ecr\.[^.]+\.amazonaws\.com(?:\.cn)?/`)
	garProjectPattern  = regexp.MustCompile(`^[a-z0-9-]+-docker\.pkg\.dev/([^/]+)/`)
	acrRegistryPattern = regexp.MustCompile(`^([a-z0-9-]+\.azurecr\.io)/`)
)

type AWSConnectionResolution struct {
	Connection *app.CloudConnection
	Implicit   bool
}

type AzureConnectionResolution struct {
	Connection *app.CloudConnection
	Implicit   bool
}

type GCPConnectionResolution struct {
	Connection *app.CloudConnection
	Implicit   bool
}

func ResolveGCPConnection(connectionName, serviceAccountEmail, identityProvider, projectID, imageURL, orgID string, connections []app.CloudConnection) (GCPConnectionResolution, error) {
	if connectionName != "" {
		matches := filterConnections(connections, func(connection app.CloudConnection) bool {
			return connection.Name == connectionName
		})
		if len(matches) != 1 {
			return GCPConnectionResolution{}, fmt.Errorf("cloud connection %q must identify exactly one connection", connectionName)
		}
		return validateGCPImageConnection(&matches[0], projectID)
	}
	if serviceAccountEmail == "" && identityProvider == "" {
		targetID, err := garProjectID(imageURL)
		if err != nil {
			return GCPConnectionResolution{}, err
		}
		matches := filterConnections(connections, func(connection app.CloudConnection) bool {
			return connection.Platform == app.CloudPlatformGCP && connection.TargetID == targetID && connection.HasCapability(app.CloudConnectionCapabilityImages)
		})
		if len(matches) == 0 {
			return GCPConnectionResolution{}, fmt.Errorf("no images-capable GCP cloud connection matches GAR project %s; set connection or service_account_email", targetID)
		}
		if len(matches) > 1 {
			return GCPConnectionResolution{}, fmt.Errorf("multiple images-capable GCP cloud connections match GAR project %s (%s); set connection explicitly", targetID, connectionNames(matches))
		}
		return GCPConnectionResolution{Connection: &matches[0]}, nil
	}
	if serviceAccountEmail == "" || identityProvider == "" {
		return GCPConnectionResolution{}, fmt.Errorf("service_account_email and workload_identity_provider must be set together")
	}
	matches := filterConnections(connections, func(connection app.CloudConnection) bool {
		return connection.Platform == app.CloudPlatformGCP && connection.Principal == serviceAccountEmail && connection.IdentityProvider == identityProvider
	})
	if len(matches) > 1 {
		return GCPConnectionResolution{}, fmt.Errorf("multiple GCP cloud connections match service_account_email %q", serviceAccountEmail)
	}
	if len(matches) == 1 {
		return validateGCPImageConnection(&matches[0], projectID)
	}
	return GCPConnectionResolution{Connection: &app.CloudConnection{
		OrgID: orgID, Name: serviceAccountEmail, Platform: app.CloudPlatformGCP, TargetID: projectID, Principal: serviceAccountEmail, IdentityProvider: identityProvider,
		AuthMode: app.CloudConnectionAuthModeLegacy, Capabilities: []app.CloudConnectionCapability{app.CloudConnectionCapabilityImages},
	}, Implicit: true}, nil
}

func validateGCPImageConnection(connection *app.CloudConnection, projectID string) (GCPConnectionResolution, error) {
	if connection.Platform != app.CloudPlatformGCP || connection.TargetID != projectID {
		return GCPConnectionResolution{}, fmt.Errorf("cloud connection %q does not target GCP project %s", connection.Name, projectID)
	}
	if !connection.HasCapability(app.CloudConnectionCapabilityImages) {
		return GCPConnectionResolution{}, fmt.Errorf("cloud connection %q does not have the images capability", connection.Name)
	}
	return GCPConnectionResolution{Connection: connection}, nil
}

func ResolveAzureConnection(connectionName, clientID, tenantID, imageURL, orgID string, connections []app.CloudConnection) (AzureConnectionResolution, error) {
	if connectionName != "" {
		matches := filterConnections(connections, func(connection app.CloudConnection) bool {
			return connection.Name == connectionName
		})
		if len(matches) != 1 {
			return AzureConnectionResolution{}, fmt.Errorf("cloud connection %q must identify exactly one connection", connectionName)
		}
		return validateAzureImageConnection(&matches[0])
	}
	if clientID == "" {
		registry, err := acrRegistry(imageURL)
		if err != nil {
			return AzureConnectionResolution{}, err
		}
		matches := filterConnections(connections, func(connection app.CloudConnection) bool {
			if connection.Platform != app.CloudPlatformAzure || !connection.HasCapability(app.CloudConnectionCapabilityImages) {
				return false
			}
			for _, candidate := range connection.Registries {
				if strings.EqualFold(candidate, registry) {
					return true
				}
			}
			return false
		})
		if len(matches) == 0 {
			return AzureConnectionResolution{}, fmt.Errorf("no images-capable Azure cloud connection is verified for ACR registry %s; set connection or client_id", registry)
		}
		if len(matches) > 1 {
			return AzureConnectionResolution{}, fmt.Errorf("multiple images-capable Azure cloud connections match ACR registry %s (%s); set connection explicitly", registry, connectionNames(matches))
		}
		return AzureConnectionResolution{Connection: &matches[0]}, nil
	}
	matches := filterConnections(connections, func(connection app.CloudConnection) bool {
		return connection.Platform == app.CloudPlatformAzure && connection.Principal == clientID
	})
	if len(matches) > 1 {
		return AzureConnectionResolution{}, fmt.Errorf("multiple Azure cloud connections match client_id %q", clientID)
	}
	if len(matches) == 1 {
		return validateAzureImageConnection(&matches[0])
	}
	return AzureConnectionResolution{Connection: &app.CloudConnection{
		OrgID: orgID, Name: clientID, Platform: app.CloudPlatformAzure, Principal: clientID, TenantID: tenantID,
		AuthMode: app.CloudConnectionAuthModeLegacy, Capabilities: []app.CloudConnectionCapability{app.CloudConnectionCapabilityImages},
	}, Implicit: true}, nil
}

func validateAzureImageConnection(connection *app.CloudConnection) (AzureConnectionResolution, error) {
	if connection.Platform != app.CloudPlatformAzure {
		return AzureConnectionResolution{}, fmt.Errorf("cloud connection %q is not an Azure connection", connection.Name)
	}
	if !connection.HasCapability(app.CloudConnectionCapabilityImages) {
		return AzureConnectionResolution{}, fmt.Errorf("cloud connection %q does not have the images capability", connection.Name)
	}
	return AzureConnectionResolution{Connection: connection}, nil
}

func ResolveAWSConnection(connectionName, roleARN, imageURL, region, orgID string, connections []app.CloudConnection) (AWSConnectionResolution, error) {
	targetID, err := ecrAccountID(imageURL)
	if err != nil {
		return AWSConnectionResolution{}, err
	}
	if connectionName != "" {
		matches := filterConnections(connections, func(connection app.CloudConnection) bool {
			return connection.Name == connectionName
		})
		if len(matches) != 1 {
			return AWSConnectionResolution{}, fmt.Errorf("cloud connection %q must identify exactly one connection", connectionName)
		}
		return validateImageConnection(&matches[0], targetID)
	}
	if roleARN != "" {
		parsed, err := arn.Parse(roleARN)
		if err != nil || parsed.Service != "iam" || parsed.AccountID != targetID || !strings.HasPrefix(parsed.Resource, "role/") {
			return AWSConnectionResolution{}, fmt.Errorf("iam_role_arn must be an IAM role in ECR account %s", targetID)
		}
		matches := filterConnections(connections, func(connection app.CloudConnection) bool {
			return connection.Platform == app.CloudPlatformAWS && connection.Principal == roleARN
		})
		if len(matches) > 1 {
			return AWSConnectionResolution{}, fmt.Errorf("multiple cloud connections match iam_role_arn %q", roleARN)
		}
		if len(matches) == 1 {
			return validateImageConnection(&matches[0], targetID)
		}
		name := parsed.Resource[strings.LastIndex(parsed.Resource, "/")+1:]
		return AWSConnectionResolution{Connection: &app.CloudConnection{OrgID: orgID, Name: name, Platform: app.CloudPlatformAWS, TargetID: targetID, Principal: roleARN, DefaultRegion: region, AuthMode: app.CloudConnectionAuthModeLegacy, RequestedCapabilities: []app.CloudConnectionCapability{app.CloudConnectionCapabilityImages}, Capabilities: []app.CloudConnectionCapability{app.CloudConnectionCapabilityImages}}, Implicit: true}, nil
	}
	matches := filterConnections(connections, func(connection app.CloudConnection) bool {
		return connection.Platform == app.CloudPlatformAWS && connection.TargetID == targetID && connection.HasCapability(app.CloudConnectionCapabilityImages)
	})
	if len(matches) == 0 {
		return AWSConnectionResolution{}, fmt.Errorf("no images-capable AWS cloud connection matches ECR account %s; set connection or iam_role_arn", targetID)
	}
	if len(matches) > 1 {
		return AWSConnectionResolution{}, fmt.Errorf("multiple images-capable AWS cloud connections match ECR account %s (%s); set connection explicitly", targetID, connectionNames(matches))
	}
	return AWSConnectionResolution{Connection: &matches[0]}, nil
}

func validateImageConnection(connection *app.CloudConnection, targetID string) (AWSConnectionResolution, error) {
	if connection.Platform != app.CloudPlatformAWS || connection.TargetID != targetID {
		return AWSConnectionResolution{}, fmt.Errorf("cloud connection %q does not target ECR account %s", connection.Name, targetID)
	}
	if !connection.HasCapability(app.CloudConnectionCapabilityImages) {
		return AWSConnectionResolution{}, fmt.Errorf("cloud connection %q does not have the images capability", connection.Name)
	}
	return AWSConnectionResolution{Connection: connection}, nil
}

func filterConnections(connections []app.CloudConnection, keep func(app.CloudConnection) bool) []app.CloudConnection {
	result := make([]app.CloudConnection, 0, len(connections))
	for _, connection := range connections {
		if keep(connection) {
			result = append(result, connection)
		}
	}
	return result
}

func ecrAccountID(imageURL string) (string, error) {
	matches := ecrAccountPattern.FindStringSubmatch(imageURL)
	if len(matches) != 2 {
		return "", fmt.Errorf("image_url must be an AWS ECR URL containing a 12-digit account ID")
	}
	return matches[1], nil
}

func garProjectID(imageURL string) (string, error) {
	matches := garProjectPattern.FindStringSubmatch(strings.ToLower(imageURL))
	if len(matches) != 2 {
		return "", fmt.Errorf("image_url must be a GAR URL in <location>-docker.pkg.dev/<project>/<repository>/... format")
	}
	return matches[1], nil
}

func acrRegistry(imageURL string) (string, error) {
	matches := acrRegistryPattern.FindStringSubmatch(strings.ToLower(imageURL))
	if len(matches) != 2 {
		return "", fmt.Errorf("image_url must be an ACR URL in <registry>.azurecr.io/... format")
	}
	return matches[1], nil
}

func connectionNames(connections []app.CloudConnection) string {
	names := make([]string, 0, len(connections))
	for _, connection := range connections {
		names = append(names, connection.Name)
	}
	return strings.Join(names, ", ")
}
