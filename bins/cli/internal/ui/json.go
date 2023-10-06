package ui

import (
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon-go"
)

func PrintJSON(data interface{}) {
	j, _ := json.Marshal(data)
	fmt.Println(string(j))
}

type jsonError struct {
	Error string `json:"error"`
}

func PrintJSONError(err error) {
	userErr, ok := nuon.ToUserError(err)
	if ok {
		PrintJSON(userErr)
		return
	}

	if nuon.IsServerError(err) {
		PrintJSON(jsonError{
			Error: defaultServerErrorMessage,
		})
		return
	}

	PrintJSON(jsonError{
		Error: defaultUnknownErrorMessage,
	})
}
