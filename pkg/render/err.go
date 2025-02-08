package render

import "fmt"

type RenderErr struct {
	Name     string
	Template string
	Err      error
}

func (r RenderErr) Error() string {
	return fmt.Sprintf("unable to render %s from %s", r.Name, r.Template)
}

func (r RenderErr) Unwrap() error {
	return r.Err
}
