package registry

func (r *Registry) Config() (interface{}, error) {
	return &r.config, nil
}
