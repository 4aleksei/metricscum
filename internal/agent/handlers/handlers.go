package handlers

import (
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/4aleksei/metricscum/internal/agent/config"
	"github.com/4aleksei/metricscum/internal/agent/service"
	"github.com/4aleksei/metricscum/internal/common/models"
)

type App struct {
	serv *service.HandlerStore
	cfg  *config.Config
}

func NewApp(store *service.HandlerStore, cfg *config.Config) *App {
	p := new(App)
	p.serv = store
	p.cfg = cfg
	return p
}

func (app *App) Run() error {

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
	server := "http://" + app.cfg.Address + "/update/"

	var plainTxtFunc = func(data string) error {

		resp, err := client.Post(server+data, "text/plain", nil)
		if err != nil {
			log.Println(err)
			return err
		}
		defer resp.Body.Close()

		_, errcoppy := io.Copy(io.Discard, resp.Body)
		if errcoppy != nil {
			log.Println(err)
			return err
		}

		return nil
	}

	var JsonModelFunc = func(data *models.Metrics) error {

		buf, err := data.JsonEncode()
		if err != nil {
			log.Println(err)
			//logger.Log.Debug("error encoding response", zap.Error(err))
			return err
		}
		resp, err := client.Post(server, "application/json", buf)
		if err != nil {
			log.Println(err)
			return err
		}
		defer resp.Body.Close()

		_, errcoppy := io.Copy(io.Discard, resp.Body)
		if errcoppy != nil {
			log.Println(err)
			return err
		}

		return nil
	}

	for {

		time.Sleep(time.Duration(app.cfg.ReportInterval) * time.Second)

		if app.cfg.ContentJson == 1 {
			app.serv.RangeMetricsJson(JsonModelFunc)
		} else {
			app.serv.RangeMetricsPlain(plainTxtFunc)
		}
	}

}
