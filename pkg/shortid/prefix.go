package shortid

import "fmt"

func GetPrefix(id string) (string, error) {
	if !IsShortID(id) {
		return "", fmt.Errorf("not a short-id")
	}

	return id[:3], nil
}
