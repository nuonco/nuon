package shortid

import (
	"reflect"

	"github.com/go-faker/faker/v4"
)

func fakeShortID(v reflect.Value) (interface{}, error) {
	fakeNanoID := NewNanoID("") //prefix=def
	return fakeNanoID, nil
}

func init() {
	_ = faker.AddProvider("shortID", fakeShortID)
}
