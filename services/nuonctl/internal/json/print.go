package json

import (
	"encoding/json"
	"fmt"
)

func Print(val interface{}) error {
	byts, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("unable to convert value to json: %w", err)
	}

	fmt.Printf("%s\n", byts)
	return nil
}
