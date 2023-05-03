package client

import "fmt"

const DefaultTokenSecretKey = "token"

func DefaultTokenSecretName(id string) string {
	return fmt.Sprintf("waypoint-bootstrap-token-%v", id)
}
