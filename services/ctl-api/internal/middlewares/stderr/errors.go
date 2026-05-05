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

// A user error is a standard user error that denotes something about the user input was not valid.
//
// ErrUser is also the canonical error type returned by signal Validate/Execute
// and by activities they call, so the same well-formed user-facing error
// travels both the HTTP path and the workflow path. The HTTP middleware
// renders Description/Code only; the workflow path additionally honors
// Directive (see step.go) to decide whether to stop, skip, or fall through
// to normal auto-retry policy. Fields carry small structured context
// (component_id, install_id, …) that the dashboard can render.
type ErrUser struct {
	Err         error
	Description string
	Code        string
	// Fields may be nil when the producer supplied no structured context.
	// Consumers MUST treat nil as equivalent to an empty map and MUST NOT
	// write into Fields without first nil-checking and allocating — the
	// map is shared across the error's lifetime (HTTP response, Temporal
	// payload, DB metadata) and producers do not deep-copy it.
	Fields    map[string]string
	Directive StepDirective
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
