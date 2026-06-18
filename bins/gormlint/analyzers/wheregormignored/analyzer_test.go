package wheregormignored_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/nuonco/nuon/bins/gormlint/analyzers/wheregormignored"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, wheregormignored.Analyzer, "example")
}
