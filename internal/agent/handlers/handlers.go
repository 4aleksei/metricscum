package handlers

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/4aleksei/metricscum/internal/agent/service"
)

func MainHTTPClient(store *service.HandlerStore) error {

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

	for {

		time.Sleep(10 * time.Second)
		server := "http://localhost:8080/update/"

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
			//fmt.Println(resp.StatusCode)
			/*for {

				bs := make([]byte, 1014)
				n, err := resp.Body.Read(bs)
				//fmt.Println(string(bs[:n]))
				if n == 0 || err != nil {
					break
				}

			}*/

			return nil
		})

	}

}
