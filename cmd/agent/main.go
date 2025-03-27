// Сервис сбора метрик и алертинга
// Приложение agent
// Собирает метрики и отправляет на server
package main

import (
	"github.com/4aleksei/metricscum/internal/agent/app"
)

func main() {
	app.SetupFX().Run()
}
