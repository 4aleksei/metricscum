// Package main - Multichecker - static linter realization.
// Custom set of  linters:
//  1. Static set of golang.org/x/tools/go/analysis/passes
//
// https://pkg.go.dev/golang.org/x/tools/go/analysis
//
//  2. staticcheck.io All analysis class SA
//
//  3. "QF1003" of staticcheck.io - Convert if/else-if chain to tagged switch
//
//  4. added go-critic linter
//
// https://go-critic.com/overview
//
//  5. added  Funlen linter
//
// https://github.com/ultraware/funlen
//
// Usage:
//  1. ./staticlint ./...
//  2. go vet -vettool=./cmd/staticlint/staticlint ./...
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	var mychecks []*analysis.Analyzer
	mychecks = multi(mychecks)

	multichecker.Main(
		mychecks...,
	)
}
