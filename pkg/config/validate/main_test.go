package validate

import (
	"os"
	"testing"

	"github.com/nuonco/nuon/pkg/publicnet/templatefetch"
)

func TestMain(m *testing.M) {
	templateOutputsClient = templatefetch.NewLocalClient()
	os.Exit(m.Run())
}
