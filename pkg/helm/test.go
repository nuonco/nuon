package helm

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jaswdr/faker"
)

const (
	fakerSeedBase  = 10
	fakerSeedWidth = 64
)

var (
	fakerSeed string
	fkr       faker.Faker
)

func init() { //nolint: gochecknoinits // removing this function will break testing suite currently
	flag.StringVar(&fakerSeed, "faker_seed", "0", "the seed timestamp for faker")
}

func TestMain(m *testing.M) {
	flag.Parse()
	seed, err := strconv.ParseInt(fakerSeed, fakerSeedBase, fakerSeedWidth)
	if err != nil {
		panic(err)
	}
	if seed == 0 {
		seed = time.Now().Unix()
	}

	fmt.Printf("Creating faker with seed: %d\n", seed)
	fkr = faker.NewWithSeed(rand.NewSource(seed))

	exitCode := m.Run()
	os.Exit(exitCode)
}
