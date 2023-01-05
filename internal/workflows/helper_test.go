package workflows

import "github.com/jaswdr/faker"

func getFakeObj[T any]() T {
	fkr := faker.New()
	var obj T
	fkr.Struct().Fill(&obj)
	return obj
}
