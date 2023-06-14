package fakers

import (
	"reflect"

	"github.com/powertoolsdev/mono/pkg/shortid"
)

func fakeShortID(v reflect.Value) (interface{}, error) {
	fakeNanoID := shortid.NewNanoID("") //prefix=def
	return fakeNanoID, nil
}
