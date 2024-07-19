package parse

import (
	"fmt"
	"os"
)

func ReadFile(filename string) ([]byte, error) {
	byts, err := os.ReadFile(filename)
	if err != nil {
		return nil, ParseErr{
			Description: fmt.Sprintf("unable to load config file %s", filename),
			Err:         err,
		}
	}

	return byts, nil
}
