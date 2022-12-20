package faker

import (
	"github.com/jaswdr/faker"
)

// GetFakeObj returns a fake version of any object
func GetFakeObj[T any]() T {
	fkr := faker.New()
	var obj T
	fkr.Struct().Fill(&obj)
	return obj
}
