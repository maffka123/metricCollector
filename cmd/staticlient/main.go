//go:build ignore
// +build ignore

/*
Package static client implements different custom static linters, such as:

- Standart linters from analysis/passes
- Linters SA and ST from staticcheck
- Couple of comunity linters
- My own linter that checks if main has os.Exit


To run them for your files:

go run cmd/staticlient/main.go ./...
*/

package main

import (
	"github.com/maffka123/metricCollector/internal/staticlient/analysispasses"
	"github.com/maffka123/metricCollector/internal/staticlient/own"
	"github.com/maffka123/metricCollector/internal/staticlient/publicanalysizers"
	"github.com/maffka123/metricCollector/internal/staticlient/staticcheckio"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {

	mychecks := staticcheckio.Staticcheckio()
	ap := analysispasses.Analysispasses()
	pu := publicanalysizers.Analyspublic()

	mychecks = append(mychecks, ap...)
	mychecks = append(mychecks, pu...)
	mychecks = append(mychecks, own.Analyzer)

	multichecker.Main(
		mychecks...,
	)
}
