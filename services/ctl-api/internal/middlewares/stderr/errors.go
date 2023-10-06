package stderr

type ErrAuthentication struct {
	Err         error
	Description string
}

func (e ErrAuthentication) Error() string {
	return e.Err.Error()
}

type ErrAuthorization struct {
	Err         error
	Description string
}

func (e ErrAuthorization) Error() string {
	return e.Err.Error()
}

// A user error is a standard user error that denotes something about the user input was not valid
type ErrUser struct {
	Err         error
	Description string
}

func (u ErrUser) Error() string {
	return u.Err.Error()
}

type ErrResponse struct {
	Error       string `json:"error"`
	UserError   bool   `json:"user_error"`
	Description string `json:"description"`
}
