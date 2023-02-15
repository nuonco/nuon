package generics

import (
	"reflect"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/powertoolsdev/go-common/shortid"
)

func registerFakerProviders() {
	_ = faker.AddProvider("shortID", func(v reflect.Value) (interface{}, error) {
		uid := uuid.New()
		shortID, err := shortid.ParseString(uid.String())
		if err != nil {
			panic(err)
		}

		return shortID, nil
	})
}

// GetFakeObj returns a faked instance of type T
// will panic on error
// meant exclusively for testing
func GetFakeObj[T any]() T {
	registerFakerProviders()

	var obj T
	err := faker.FakeData(&obj)
	if err != nil {
		panic(err)
	}
	return obj
}
