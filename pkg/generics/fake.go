package generics

import (
	"github.com/go-faker/faker/v4"

	"github.com/nuonco/nuon/pkg/shortid"
)

func GetFakeObj[T any]() T {
	shortid.RegisterFakes()
	var obj T
	err := faker.FakeData(&obj)
	if err != nil {
		panic(err)
	}
	return obj
}
