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

// A not ready error
type ErrNotReady struct {
	Err         error
	Description string
}

func (u ErrNotReady) Error() string {
	return u.Err.Error()
}

type ErrNotFound struct {
	Err         error
	Description string
}

func (e ErrNotFound) Error() string {
	return e.Err.Error()
}

type ErrResponse struct {
	Error       string `json:"error,omitzero"`
	UserError   bool   `json:"user_error,omitzero"`
	Description string `json:"description,omitzero"`
}

type ErrSystem struct {
	Err         error
	Description string
}

func (e ErrSystem) Error() string {
	return e.Err.Error()
}

type ErrInvalidRequest struct {
	Err error
}

func (e ErrInvalidRequest) Error() string {
	return e.Err.Error()
}

func (e ErrInvalidRequest) Unwrap() error {
	return e.Err
}
