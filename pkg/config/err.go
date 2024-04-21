package config

type ErrConfig struct {
	Description string
	Err         error
}

func (e ErrConfig) Error() string {
	return e.Err.Error()
}
