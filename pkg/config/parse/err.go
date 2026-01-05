package parse

type ParseErr struct {
	Filename    string
	Description string
	Err         error
}

func (p ParseErr) Error() string {
	if p.Filename != "" {
		return p.Filename + ": " + p.Description
	}
	return p.Description
}
