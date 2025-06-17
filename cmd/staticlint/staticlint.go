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
	"fmt"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func printVersion() {
	fmt.Println("Build version: ", buildVersion)
	fmt.Println("Build date: ", buildDate)
	fmt.Println("Build commit: ", buildCommit)
}

func main() {
	printVersion()
	var mychecks []*analysis.Analyzer
	mychecks = multi(mychecks)

	multichecker.Main(
		mychecks...,
	)
}
