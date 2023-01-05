package services

import (
	"github.com/jaswdr/faker"
)

func getFakeObj[T any]() T {
	var obj T
	fkr := faker.New()
	fkr.Struct().Fill(&obj)
	return obj
}

func toPtr[T any](t T) *T {
	return &t
}
