package handlers

import (
	"bytes"
	"compress/gzip"
	"context"

	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/4aleksei/metricscum/internal/agent/config"
	"github.com/4aleksei/metricscum/internal/agent/service"
	"github.com/4aleksei/metricscum/internal/common/logger"
	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/utils"
	"go.uber.org/zap"
)

type App struct {
	serv   *service.HandlerStore
	cfg    *config.Config
	l      *logger.Logger
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

const (
	textHTMLContent        string = "text/html"
	textPlainContent       string = "text/plain"
	applicationJSONContent string = "application/json"
	gzipContent            string = "gzip"
)

func NewApp(store *service.HandlerStore, l *logger.Logger, cfg *config.Config) *App {
	p := new(App)
	p.serv = store
	p.cfg = cfg
	p.l = l
	return p
}

func (app *App) newPPostReq(ctx context.Context, client *http.Client, server string, requestBody io.Reader) error {
	req, err := http.NewRequestWithContext(ctx, "POST", server, requestBody)
	if err != nil {
		app.l.L.Debug("Error create new request", zap.Error(err))
		return err
	}
	req.Header.Set("Content-Type", textPlainContent)
	resp, err := client.Do(req)
	if err != nil {
		app.l.L.Debug("Error send request", zap.Error(err))
		return err
	}
	defer resp.Body.Close()
	_, errcoppy := io.Copy(io.Discard, resp.Body)
	if errcoppy != nil {
		app.l.L.Debug("Error read response", zap.Error(err))
		return err
	}
	return nil
}

func (app *App) newJPostReq(ctx context.Context, client *http.Client, server string, requestBody io.Reader) error {
	req, err := http.NewRequestWithContext(ctx, "POST", server, requestBody)
	if err != nil {
		app.l.L.Debug("Error create new request", zap.Error(err))
		return err
	}
	req.Header.Set("Accept-Encoding", gzipContent)
	req.Header.Set("Content-Encoding", gzipContent)
	req.Header.Set("Content-Type", applicationJSONContent)
	req.Header.Set("Accept", applicationJSONContent)
	resp, err := client.Do(req)
	if err != nil {
		app.l.L.Debug("Error send request", zap.Error(err))
		return err
	}
	defer resp.Body.Close()
	_, errcoppy := io.Copy(io.Discard, resp.Body)
	if errcoppy != nil {
		app.l.L.Debug("Error read response", zap.Error(err))
		return err
	}
	return nil
}

func (app *App) Start(ctx context.Context) error {
	ctxCancel, cancel := context.WithCancel(context.Background())
	app.cancel = cancel
	app.wg = sync.WaitGroup{}
	app.wg.Add(1)
	go app.run(ctxCancel)
	return nil
}

func (app *App) Stop(ctx context.Context) error {
	app.cancel()
	app.wg.Wait()
	return nil
}

func (app *App) run(ctx context.Context) {
	defer app.wg.Done()
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 2 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 2 * time.Second,
	}
	var client = &http.Client{
		Transport: netTransport,
	}
	server := "http://" + app.cfg.Address + "/update/"
	var plainTxtFunc = func(ctx context.Context, data string) error {
		err := utils.RetryAction(ctx, utils.RetryTimes(), func(ctx context.Context) error {
			return app.newPPostReq(ctx, client, server, http.NoBody)
		})

		if err != nil {
			return err
		}
		return nil
	}

	var jsonModelFunc = func(ctx context.Context, data *models.Metrics) error {
		var requestBody bytes.Buffer
		gz := gzip.NewWriter(&requestBody)
		err := data.JSONEncodeBytes(gz)
		gz.Close()
		if err != nil {
			app.l.L.Debug("error json encoding", zap.Error(err))
			return err
		}
		err = utils.RetryAction(ctx, utils.RetryTimes(), func(ctx context.Context) error {
			return app.newJPostReq(ctx, client, server, &requestBody)
		})
		if err != nil {
			return err
		}
		return nil
	}
	var jsonModelSFunc = func(ctx context.Context, data []models.Metrics) error {
		var requestBody bytes.Buffer
		gz := gzip.NewWriter(&requestBody)
		err := models.JSONSEncodeBytes(gz, data)
		gz.Close()
		if err != nil {
			app.l.L.Debug("error json encoding", zap.Error(err))
			return err
		}

		serverupdates := "http://" + app.cfg.Address + "/updates/"
		err = utils.RetryAction(ctx, utils.RetryTimes(), func(ctx context.Context) error {
			return app.newJPostReq(ctx, client, serverupdates, &requestBody)
		})

		if err != nil {
			return err
		}
		return nil
	}

	for {
		select {
		case <-ctx.Done():

			return
		default:
			utils.SleepContext(ctx, time.Duration(app.cfg.ReportInterval)*time.Second)

			if app.cfg.ContentJSON {
				if app.cfg.ContentBatch {
					_ = app.serv.RangeMetricsJSONS(ctx, jsonModelSFunc)
				} else {
					_ = app.serv.RangeMetricsJSON(ctx, jsonModelFunc)
				}
			} else {
				_ = app.serv.RangeMetrics(ctx, plainTxtFunc)
			}
		}
	}
}
