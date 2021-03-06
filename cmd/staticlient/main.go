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
	"github.com/bkielbasa/cyclop/pkg/analyzer"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {

	mychecks := Staticcheckio()
	ap := Analysispasses()
	pu := Analyspublic()

	mychecks = append(mychecks, ap...)
	mychecks = append(mychecks, pu...)
	mychecks = append(mychecks, Analyzer)

	multichecker.Main(
		mychecks...,
	)
}

func Analyspublic() []*analysis.Analyzer {
	var mychecks []*analysis.Analyzer
	mychecks = append(mychecks, bodyclose.Analyzer)
	mychecks = append(mychecks, analyzer.NewAnalyzer())

	return mychecks
}

var Analyzer = &analysis.Analyzer{
	Name: "checkosexit",
	Doc:  "Checks if os.Exit was used in main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name == "main" {
			ast.Inspect(file, func(node ast.Node) bool {

				if funcCall, ok := node.(*ast.CallExpr); ok {
					if fun, ok := funcCall.Fun.(*ast.SelectorExpr); ok {
						if fun.Sel.Name == "Exit" {
							if p, ok := fun.X.(*ast.Ident); ok {
								if p.Name == "os" {
									pass.Reportf(p.NamePos, "os.Exit in main")
								}
							}

						}
					}
				}

				return true
			})

		}
	}
	return nil, nil
}

func Analysispasses() []*analysis.Analyzer {
	var mychecks []*analysis.Analyzer
	mychecks = append(mychecks, printf.Analyzer)
	mychecks = append(mychecks, shadow.Analyzer)
	mychecks = append(mychecks, structtag.Analyzer)

	return mychecks
}

func Staticcheckio() []*analysis.Analyzer {

	var mychecks []*analysis.Analyzer
	// SA analyzers
	for _, v := range staticcheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}
	// ST analyzers
	for _, v := range stylecheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}

	return mychecks
}
