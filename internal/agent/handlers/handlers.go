package handlers

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/4aleksei/metricscum/internal/agent/service"
)

func MainHTTPClient(store *service.HandlerStore, addr string, reportInterval uint) error {

	//client := http.Client{}

	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 2 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 2 * time.Second,
	}
	var client = &http.Client{
		Timeout:   time.Second * 5,
		Transport: netTransport,
	}
	server := "http://" + addr + "/update/"
	for {

		time.Sleep(time.Duration(reportInterval) * time.Second)

		service.RangeMetrics(store.Store, func(data string) error {

			resp, err := client.Post(server+data, "Content-Type: text/plain", nil)
			if err != nil {
				fmt.Println(err)
				return err
			}
			defer resp.Body.Close()

			_, errcoppy := io.Copy(io.Discard, resp.Body)
			if errcoppy != nil {
				fmt.Println(err)
				return err
			}

			return nil
		})

	}

}
