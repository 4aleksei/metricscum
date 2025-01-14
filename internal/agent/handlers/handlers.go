package handlers

import (
	"bytes"
	"compress/gzip"
	"context"
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

const (
	textHTMLContent        string = "text/html"
	applicationJSONContent string = "application/json"
	gzipContent            string = "gzip"
)

func NewApp(store *service.HandlerStore, cfg *config.Config) *App {
	p := new(App)
	p.serv = store
	p.cfg = cfg
	return p
}

func newPostReq(ctx context.Context, client *http.Client, server string, requestBody io.Reader) error {
	req, err := http.NewRequestWithContext(ctx, "POST", server, requestBody)
	if err != nil {
		log.Println(err)
		return err
	}
	req.Header.Set("Accept-Encoding", gzipContent)
	req.Header.Set("Content-Encoding", gzipContent)
	req.Header.Set("Content-Type", applicationJSONContent)
	req.Header.Set("Accept", applicationJSONContent)
	resp, err := client.Do(req)
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
		ctx := context.Background()
		req, err := http.NewRequestWithContext(ctx, "POST", server, http.NoBody)
		if err != nil {
			log.Println(err)
			return err
		}
		req.Header.Set("Content-Type", "text/plain")
		resp, err := client.Do(req)
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

	var JSONModelFunc = func(data *models.Metrics) error {
		var requestBody bytes.Buffer
		gz := gzip.NewWriter(&requestBody)
		err := data.JSONEncodeBytes(gz)
		gz.Close()
		if err != nil {
			log.Println(err)
			return err
		}
		ctx := context.Background()
		err = newPostReq(ctx, client, server, &requestBody)
		if err != nil {
			return err
		}
		return nil
	}
	var JSONModelSFunc = func(data *[]models.Metrics) error {
		var requestBody bytes.Buffer
		gz := gzip.NewWriter(&requestBody)
		err := models.JSONSEncodeBytes(gz, data)
		gz.Close()
		if err != nil {
			log.Println(err)
			return err
		}
		ctx := context.Background()
		serverupdates := "http://" + app.cfg.Address + "/updates/"
		err = newPostReq(ctx, client, serverupdates, &requestBody)
		if err != nil {
			return err
		}
		return nil
	}

	for {
		time.Sleep(time.Duration(app.cfg.ReportInterval) * time.Second)
		if app.cfg.ContentJSON {
			if app.cfg.ContentBatch {
				_ = app.serv.RangeMetricsJSONS(JSONModelSFunc)
			} else {
				_ = app.serv.RangeMetricsJSON(JSONModelFunc)
			}
		} else {
			_ = app.serv.RangeMetrics(plainTxtFunc)
		}
	}
}
