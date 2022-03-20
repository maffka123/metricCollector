// Package own is my personal check which prohibits usage of os.Exit in function main of package main
package own

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

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
