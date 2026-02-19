package stderr

type ErrAuthentication struct {
	Err         error
	Description string
}

func (e ErrAuthentication) Error() string {
	return e.Err.Error()
}

func (e ErrAuthentication) Unwrap() error {
	return e.Err
}

type ErrAuthorization struct {
	Err         error
	Description string
}

func (e ErrAuthorization) Error() string {
	return e.Err.Error()
}

func (e ErrAuthorization) Unwrap() error {
	return e.Err
}

// A user error is a standard user error that denotes something about the user input was not valid
type ErrUser struct {
	Err         error
	Description string
	Code        string
}

func (u ErrUser) Error() string {
	return u.Err.Error()
}

func (u ErrUser) Unwrap() error {
	return u.Err
}

// A not ready error
type ErrNotReady struct {
	Err         error
	Description string
}

func (u ErrNotReady) Error() string {
	return u.Err.Error()
}

func (u ErrNotReady) Unwrap() error {
	return u.Err
}

type ErrNotFound struct {
	Err         error
	Description string
}

func (e ErrNotFound) Error() string {
	return e.Err.Error()
}

func (e ErrNotFound) Unwrap() error {
	return e.Err
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

func (e ErrSystem) Unwrap() error {
	return e.Err
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

func NewInvalidRequest(err error) ErrInvalidRequest {
	return ErrInvalidRequest{Err: err}
}

type ErrConflict struct {
	Err         error
	Description string
}

func (e ErrConflict) Error() string {
	return e.Err.Error()
}

func (e ErrConflict) Unwrap() error {
	return e.Err
}
