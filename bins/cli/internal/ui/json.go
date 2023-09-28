package ui

import (
	"encoding/json"
	"fmt"
)

func printJSON(data string) {
	j, _ := json.Marshal(data)
	fmt.Println(string(j))
}

func printJSONError(err error) {
	printJSON(fmt.Sprintf("{ \"error\": %s }", err))
}
