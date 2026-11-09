package validate

import (
	"net/http"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	templateOutputsClient = &http.Client{Timeout: 30 * time.Second}
	os.Exit(m.Run())
}
