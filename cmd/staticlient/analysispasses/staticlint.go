// Package analysispasses implements some checks from analysis/passes lib
package analysispasses

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
)

func Analysispasses() []*analysis.Analyzer {
	var mychecks []*analysis.Analyzer
	mychecks = append(mychecks, printf.Analyzer)
	mychecks = append(mychecks, shadow.Analyzer)
	mychecks = append(mychecks, structtag.Analyzer)

	return mychecks
}
