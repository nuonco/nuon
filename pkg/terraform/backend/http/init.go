package http

import "context"

func (s *http) Init(context.Context) error {
	return s.startProxy()
}
