package ui

import (
	"fmt"

	"github.com/nuonco/nuon-go"
)

func PrintError(err error) {
	userErr, ok := nuon.ToUserError(err)
	if ok {
		fmt.Println(userErr.Description)
		return
	}

	if nuon.IsServerError(err) {
		fmt.Println(defaultServerErrorMessage)
		return
	}

	fmt.Println(defaultUnknownErrorMessage)
}
