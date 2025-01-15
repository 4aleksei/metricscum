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
	"github.com/4aleksei/metricscum/internal/common/utils"
)

type App struct {
	serv *service.HandlerStore
	cfg  *config.Config
}

const (
	textHTMLContent        string = "text/html"
	textPlainContent       string = "text/plain"
	applicationJSONContent string = "application/json"
	gzipContent            string = "gzip"
)

func NewApp(store *service.HandlerStore, cfg *config.Config) *App {
	p := new(App)
	p.serv = store
	p.cfg = cfg
	return p
}

func newPPostReq(ctx context.Context, client *http.Client, server string, requestBody io.Reader) error {
	req, err := http.NewRequestWithContext(ctx, "POST", server, requestBody)
	if err != nil {
		log.Println(err)
		return err
	}
	req.Header.Set("Content-Type", textPlainContent)
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

func newJPostReq(ctx context.Context, client *http.Client, server string, requestBody io.Reader) error {
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

func (app *App) Run(ctxMAIN context.Context) error {
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
	var plainTxtFunc = func(ctx context.Context, data string) error {
		err := utils.RetryAction(ctx, utils.RetryTimes(), func(ctx context.Context) error {
			return newPPostReq(ctx, client, server, http.NoBody)
		})

		if err != nil {
			return err
		}
		return nil
	}

	var JSONModelFunc = func(ctx context.Context, data *models.Metrics) error {
		var requestBody bytes.Buffer
		gz := gzip.NewWriter(&requestBody)
		err := data.JSONEncodeBytes(gz)
		gz.Close()
		if err != nil {
			log.Println(err)
			return err
		}
		err = utils.RetryAction(ctx, utils.RetryTimes(), func(ctx context.Context) error {
			return newJPostReq(ctx, client, server, &requestBody)
		})
		if err != nil {
			return err
		}
		return nil
	}
	var JSONModelSFunc = func(ctx context.Context, data *[]models.Metrics) error {
		var requestBody bytes.Buffer
		gz := gzip.NewWriter(&requestBody)
		err := models.JSONSEncodeBytes(gz, data)
		gz.Close()
		if err != nil {
			log.Println(err)
			return err
		}

		serverupdates := "http://" + app.cfg.Address + "/updates/"
		err = utils.RetryAction(ctx, utils.RetryTimes(), func(ctx context.Context) error {
			return newJPostReq(ctx, client, serverupdates, &requestBody)
		})

		if err != nil {
			return err
		}
		return nil
	}

	for {
		time.Sleep(time.Duration(app.cfg.ReportInterval) * time.Second)
		if app.cfg.ContentJSON {
			if app.cfg.ContentBatch {
				_ = app.serv.RangeMetricsJSONS(ctxMAIN, JSONModelSFunc)
			} else {
				_ = app.serv.RangeMetricsJSON(ctxMAIN, JSONModelFunc)
			}
		} else {
			_ = app.serv.RangeMetrics(ctxMAIN, plainTxtFunc)
		}
	}
}
