package fakers

import (
	"reflect"

	"github.com/powertoolsdev/mono/pkg/common/shortid"
)

func fakeShortID(v reflect.Value) (interface{}, error) {
	fakeNanoID, _ := shortid.NewNanoID("") //prefix=def
	return fakeNanoID, nil
}
