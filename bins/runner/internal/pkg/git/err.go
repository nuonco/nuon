package git

import "fmt"

type CloneErr struct {
	Url string
	Ref string
	Err error
}

func (c CloneErr) Error() string {
	return fmt.Sprintf("unable to clone repo %s with ref %s - %s", c.Url, c.Ref, c.Err)
}

func (c CloneErr) Unwrap() error {
	return c.Err
}
