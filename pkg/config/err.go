package config

type ErrConfig struct {
	Description string
	Err         error
}

func (e ErrConfig) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	return e.Description
}
