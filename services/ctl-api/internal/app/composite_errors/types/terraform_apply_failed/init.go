package terraform_apply_failed

import (
	composite_error "github.com/nuonco/nuon/pkg/composite_error"
	"github.com/nuonco/nuon/pkg/composite_error/catalog"
)

func init() {
	catalog.RegisterType(Type, func() composite_error.CompositeError {
		return &Error{}
	})
	catalog.RegisterParser(Parser{})
}
