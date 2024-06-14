package permissions

import "fmt"

type NoAccessError struct {
	Permission Permission
	ObjectID   string
}

func (n NoAccessError) Error() string {
	return fmt.Sprintf("%s on %s is not authorized", n.Permission, n.ObjectID)
}
