package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/nuonco/nuon/bins/gormlint/analyzers/hardcodedtablename"
	"github.com/nuonco/nuon/bins/gormlint/analyzers/missingcontext"
	"github.com/nuonco/nuon/bins/gormlint/analyzers/rawsqlwhere"
	"github.com/nuonco/nuon/bins/gormlint/analyzers/unboundedpreload"
	"github.com/nuonco/nuon/bins/gormlint/analyzers/unboundedquery"
	"github.com/nuonco/nuon/bins/gormlint/analyzers/wheregormignored"
	"github.com/nuonco/nuon/bins/gormlint/analyzers/wherezerovalue"
)

func main() {
	multichecker.Main(
		wheregormignored.Analyzer,
		wherezerovalue.Analyzer,
		rawsqlwhere.Analyzer,
		hardcodedtablename.Analyzer,
		unboundedpreload.Analyzer,
		unboundedquery.Analyzer,
		missingcontext.Analyzer,
	)
}
