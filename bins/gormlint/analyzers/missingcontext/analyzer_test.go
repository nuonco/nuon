package missingcontext_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/nuonco/nuon/bins/gormlint/analyzers/missingcontext"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, missingcontext.Analyzer, "example")
}
