// Package staticcheckio implements all SA and ST checks from staticcheck
package staticcheckio

import (
	"golang.org/x/tools/go/analysis"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

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
