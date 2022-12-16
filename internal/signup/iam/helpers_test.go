package iam

import (
	"log"

	"github.com/go-faker/faker/v4"
)

func getFakeObj[T any]() T {
	var obj T
	err := faker.FakeData(&obj)
	if err != nil {
		log.Fatalf("unable to create fake obj: %s", err)
	}
	return obj
}
