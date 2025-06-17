// Metrics Alerting Service
// Application agent
// Gathering and sending metrics
package main

import (
	"fmt"

	"github.com/4aleksei/metricscum/internal/agent/app"
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
	app.SetupFX().Run()
}
