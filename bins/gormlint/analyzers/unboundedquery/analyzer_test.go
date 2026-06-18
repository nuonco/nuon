package unboundedquery_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/nuonco/nuon/bins/gormlint/analyzers/unboundedquery"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, unboundedquery.Analyzer, "example")
}
