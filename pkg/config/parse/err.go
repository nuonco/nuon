package parse

import "fmt"

type ParseErr struct {
	Filename    string
	Description string
	Err         error
}

func (p ParseErr) Error() string {
	description := p.Description
	if p.Err != nil {
		description = fmt.Sprintf("%s: %v", description, p.Err)
	}
	if p.Filename != "" {
		return p.Filename + ": " + description
	}
	return description
}

func (p ParseErr) Unwrap() error {
	return p.Err
}
