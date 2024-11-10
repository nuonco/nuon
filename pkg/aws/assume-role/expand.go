package iam

import (
	"fmt"
	"strings"
)

func ExpandRoleName(fullARN string, roleName string) string {
	pieces := strings.SplitN(fullARN, "/", 2)
	return fmt.Sprintf("%s/%s", pieces[0], roleName)
}
