package proto

import (
	"encoding/json"
	"fmt"
)

func Print(msg interface{}) error {
	byts, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("unable to marshal component: %w", err)
	}

	fmt.Printf("%s\n", byts)
	return nil
}
