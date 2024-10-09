package sync

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

func (s *sync) getChecksum(v interface{}) (string, error) {
	// Marshal the struct to JSON
	jsonData, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	// Create a SHA256 hash of the JSON data
	hash := sha256.New()
	_, err = hash.Write(jsonData)
	if err != nil {
		return "", err
	}

	// Get the checksum as a byte slice
	checksum := hash.Sum(nil)

	// Convert the checksum to a hexadecimal string
	return fmt.Sprintf("%x", checksum), nil
}
