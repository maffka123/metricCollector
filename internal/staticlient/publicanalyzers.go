// Package staticlient implements couple of public code analyzers
package staticlient

import (
	"github.com/bkielbasa/cyclop/pkg/analyzer"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
)

func Analyspublic() []*analysis.Analyzer {
	var mychecks []*analysis.Analyzer
	mychecks = append(mychecks, bodyclose.Analyzer)
	mychecks = append(mychecks, analyzer.NewAnalyzer())

	return mychecks
}
