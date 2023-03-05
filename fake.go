package generics

import (
	"github.com/go-faker/faker/v4"
	"github.com/powertoolsdev/go-generics/fakers"
)

// GetFakeObj returns a faked instance of type T
// will panic on error
// meant exclusively for testing
func GetFakeObj[T any]() T {
	fakers.Register()

	var obj T
	err := faker.FakeData(&obj)
	if err != nil {
		panic(err)
	}
	return obj
}
