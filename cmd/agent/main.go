// Metrics Alerting Service
// Application agent
// Gathering and sending metrics
package main

import (
	"github.com/4aleksei/metricscum/internal/agent/app"
)

func main() {
	app.SetupFX().Run()
}
