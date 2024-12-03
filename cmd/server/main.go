package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/4aleksei/metricscum/internal/common/repository"
	"github.com/4aleksei/metricscum/internal/server/handlers"
	"github.com/4aleksei/metricscum/internal/server/service"
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
	flag.Parse()

	store := repository.NewStore()
	metricsService := service.NewHandlerStore(store)

	return handlers.Serve(metricsService, addr.String())
}
