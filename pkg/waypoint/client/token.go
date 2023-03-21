package client

import "fmt"

func DefaultTokenSecretName(id string) string {
	return fmt.Sprintf("waypoint-bootstrap-token-%v", id)
}
