package static

import "context"

func (v *vars) Init(context.Context) error {
	return nil
}

func (v *vars) GetEnv(context.Context) (map[string]string, error) {
	return nil, nil
}

func (v *vars) GetFile(context.Context) ([]byte, error) {
	return []byte(nil), nil
}
