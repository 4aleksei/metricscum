package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/4aleksei/metricscum/internal/agent/gather"
	"github.com/4aleksei/metricscum/internal/agent/handlers"
	"github.com/4aleksei/metricscum/internal/agent/service"
	"github.com/4aleksei/metricscum/internal/common/repository"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

type NetAddress struct {
	Host string
	Port int
}

func (o *NetAddress) String() string {
	return fmt.Sprintf("%s:%d", o.Host, o.Port)
}

func (o *NetAddress) Set(flagValue string) error {
	valstr := strings.Split(flagValue, ":")
	o.Host = valstr[0]
	o.Port, _ = strconv.Atoi(valstr[1])
	return nil
}

func run() error {
	//	cfg := config.GetConfig()
	addr := new(NetAddress)
	addr.Host = "localhost"
	addr.Port = 8080

	_ = flag.Value(addr)

	flag.Var(addr, "a", "Net address host:port")

	reportInterval := flag.Uint("r", 10, "reportInterval")
	pollInterval := flag.Uint("p", 2, "pollInterval")

	flag.Parse()

	addtenv := os.Getenv("ADDRESS")
	fmt.Println(addtenv)

	store := repository.NewStoreMux()
	metricsService := service.NewHandlerStore(store)
	go gather.MainGather(metricsService, *pollInterval)
	return handlers.MainHTTPClient(metricsService, addr.String(), *reportInterval)
}
