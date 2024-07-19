package parse

type ParseErr struct {
	Description string
	Err         error
}

func (p ParseErr) Error() string {
	return p.Description
}
