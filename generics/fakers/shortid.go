package fakers

import (
	"reflect"

	"github.com/powertoolsdev/go-common/shortid"
)

func fakeShortID(v reflect.Value) (interface{}, error) {
	return shortid.New(), nil
}
