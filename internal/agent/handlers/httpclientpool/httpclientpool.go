// Package httpclientpool - Integration tests
package httpclientpool

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"net"
	"time"

	"bytes"
	"compress/gzip"
	"encoding/hex"

	"github.com/4aleksei/metricscum/internal/agent/config"
	"github.com/4aleksei/metricscum/internal/common/job"

	"github.com/4aleksei/metricscum/internal/agent/handlers/httpclientpool/httpaes"
	"github.com/4aleksei/metricscum/internal/common/middleware/hmacsha256"
	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/utils"
)

const (
	textHTMLContent        string = "text/html"
	textPlainContent       string = "text/plain"
	applicationJSONContent string = "application/json"
	gzipContent            string = "gzip"
)

var (
	ErrReadDone   = errors.New("read done")
	ErrChanClosed = errors.New("closed chan")
)

type (
	PoolHandler struct {
		cfg         *config.Config
		clients     []clientInstance
		WorkerCount int
		publicKey   *rsa.PublicKey
	}

	functioExec func(context.Context, *sync.WaitGroup, *agentClient,
		<-chan job.Job, chan<- job.Result, *config.Config, *rsa.PublicKey)
	clientInstance struct {
		execFn    functioExec
		client    *agentClient
		cfg       *config.Config
		publicKey *rsa.PublicKey
	}

	agentClient struct {
		client    *http.Client
		localAddr string
	}
)

func NewHandler(cfg *config.Config) *PoolHandler {
	p := new(PoolHandler)
	p.WorkerCount = int(cfg.RateLimit)
	p.clients = make([]clientInstance, p.WorkerCount)
	p.cfg = cfg

	if cfg.PublicKeyFile != "" {
		pub, err := httpaes.LoadPublicKey(cfg.PublicKeyFile)
		if err != nil {
			fmt.Println("Error load public key")
		} else {
			p.publicKey = pub

		}
	}

	for i := 0; i < p.WorkerCount; i++ {
		p.clients[i] = *newClientInstance(cfg, p.publicKey)
	}
	return p
}

func newClientInstance(cfg *config.Config, p *rsa.PublicKey) *clientInstance {
	return &clientInstance{
		execFn:    poolOptions(cfg),
		client:    newClient(),
		cfg:       cfg,
		publicKey: p,
	}
}

func poolOptions(cfg *config.Config) functioExec {
	if cfg.ContentJSON {
		if cfg.ContentBatch > 0 {
			return workerJSONBatch
		} else {
			return workerJSON
		}
	} else {
		return workerPlain
	}
}

func newClient() *agentClient {
	connection := &net.Dialer{
		Timeout: 2 * time.Second,
	}
	agclient := &agentClient{}

	var netTransport = &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			conn, err := connection.Dial(network, addr)
			if err == nil {
				agclient.localAddr = conn.LocalAddr().(*net.TCPAddr).IP.String()
			}
			return conn, err
		},
		TLSHandshakeTimeout: 2 * time.Second,
	}

	agclient.client = &http.Client{
		Transport: netTransport,
	}
	return agclient
}

func workerJSONBatch(ctx context.Context, wg *sync.WaitGroup, client *agentClient,

	jobs <-chan job.Job, results chan<- job.Result, cfg *config.Config, pub *rsa.PublicKey) {
	defer wg.Done()
	server := "http://" + cfg.Address + "/updates/"
	for j := range jobs {
		select {
		case <-ctx.Done():
			return
		default:

			err := jsonModelSFunc(ctx, server, client, j.Value, cfg.Key, pub)
			if err != nil && errors.Is(err, context.Canceled) {
				return
			}
			var res = job.Result{
				Err: err,
				ID:  j.ID,
			}
			results <- res
		}
	}
}

func workerJSON(ctx context.Context, wg *sync.WaitGroup, client *agentClient,
	jobs <-chan job.Job, results chan<- job.Result, cfg *config.Config, pub *rsa.PublicKey) {
	defer wg.Done()
	server := "http://" + cfg.Address + "/update/"
	for j := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
			err := jsonModelFunc(ctx, server, client, &j.Value[0], cfg.Key, pub)
			if err != nil && errors.Is(err, context.Canceled) {
				return
			}
			var res = job.Result{
				Err: err,
				ID:  j.ID,
			}
			results <- res
		}
	}
}

func workerPlain(ctx context.Context, wg *sync.WaitGroup, client *agentClient,
	jobs <-chan job.Job, results chan<- job.Result, cfg *config.Config, pub *rsa.PublicKey) {
	defer wg.Done()
	server := "http://" + cfg.Address + "/update/"
	for j := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
			data := j.Value[0].MType + "/" + j.Value[0].ID + "/" + j.Value[0].ConvertMetricToValue()
			err := plainTxtFunc(ctx, client, server, data)
			if err != nil && errors.Is(err, context.Canceled) {
				return
			}
			var res = job.Result{
				Err: err,
				ID:  j.ID,
			}
			results <- res
		}
	}
}

func (p *PoolHandler) GracefulStop() {
	for _, v := range p.clients {
		v.client.client.CloseIdleConnections()
	}
}

func (p *PoolHandler) StartPool(ctx context.Context, jobs chan job.Job, results chan job.Result, wg *sync.WaitGroup) {
	for i := 0; i < p.WorkerCount; i++ {
		wg.Add(1)
		go p.clients[i].execFn(ctx, wg, p.clients[i].client, jobs, results, p.cfg, p.clients[i].publicKey)
	}
}

func plainTxtFunc(ctx context.Context, client *agentClient, server, data string) error {
	err := utils.RetryAction(ctx, utils.RetryTimes(), func(ctx context.Context) error {
		return newPPostReq(ctx, client, server+data, http.NoBody)
	})
	if err != nil {
		return err
	}
	return nil
}

func jsonModelFunc(ctx context.Context, server string, client *agentClient, data *models.Metrics, cfgkey string, pub *rsa.PublicKey) error {
	var requestBody bytes.Buffer
	var aeskey string
	var key string
	var twr io.Writer

	gz := gzip.NewWriter(&requestBody)
	var hmac *hmacsha256.HmacWriter
	if cfgkey != "" {
		hmac = hmacsha256.NewWriter(gz, []byte(cfgkey))
		twr = hmac
	} else {
		twr = gz
	}

	if pub != nil {
		aestwr, err := httpaes.NewWriter(twr, pub)
		if err != nil {
			return err
		}
		twr = aestwr
		aeskey = aestwr.GetKey()
	}

	err := data.JSONEncodeBytes(twr)
	if err != nil {
		return err
	}
	gz.Close()

	if hmac != nil {
		key = hex.EncodeToString(hmac.GetSig())
	}

	err = utils.RetryAction(ctx, utils.RetryTimes(), func(ctx context.Context) error {
		return newJPostReq(ctx, client, server, &requestBody, key, aeskey)
	})
	if err != nil {
		return err
	}
	return nil
}

func jsonModelSFunc(ctx context.Context, server string, client *agentClient, data []models.Metrics, cfgkey string, pub *rsa.PublicKey) error {
	var requestBody bytes.Buffer
	var key string
	var twr io.Writer
	var aeskey string
	gz := gzip.NewWriter(&requestBody)
	var hmac *hmacsha256.HmacWriter

	if cfgkey != "" {
		hmac = hmacsha256.NewWriter(gz, []byte(cfgkey))
		twr = hmac
	} else {
		twr = gz
	}

	if pub != nil {
		aestwr, err := httpaes.NewWriter(twr, pub)
		if err != nil {
			return err
		}
		twr = aestwr
		aeskey = aestwr.GetKey()
	}

	err := models.JSONSEncodeBytes(twr, data)
	if err != nil {
		return err
	}
	gz.Close()

	if hmac != nil {
		key = hex.EncodeToString(hmac.GetSig())
	}

	err = utils.RetryAction(ctx, utils.RetryTimes(), func(ctx context.Context) error {
		return newJPostReq(ctx, client, server, &requestBody, key, aeskey)
	})
	if err != nil {
		return err
	}
	return nil
}

func newPPostReq(ctx context.Context, client *agentClient, server string, requestBody io.Reader) error {
	req, err := http.NewRequestWithContext(ctx, "POST", server, requestBody)

	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", textPlainContent)

	if client.localAddr != "" {
		req.Header.Set("X-Real-IP", client.localAddr)
	}

	resp, err := client.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, errcoppy := io.Copy(io.Discard, resp.Body)

	if errcoppy != nil {
		return err
	}
	return nil
}

func newJPostReq(ctx context.Context, client *agentClient, server string, requestBody io.Reader, key string, aeskey string) error {
	req, err := http.NewRequestWithContext(ctx, "POST", server, requestBody)

	if err != nil {
		return err
	}

	if key != "" {
		req.Header.Set("HashSHA256", key)
	}

	if aeskey != "" {
		req.Header.Set("AES-256", aeskey)
	}

	if client.localAddr != "" {
		req.Header.Set("X-Real-IP", client.localAddr)
	}

	req.Header.Set("Accept-Encoding", gzipContent)
	req.Header.Set("Content-Encoding", gzipContent)
	req.Header.Set("Content-Type", applicationJSONContent)
	req.Header.Set("Accept", applicationJSONContent)
	resp, err := client.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, errcoppy := io.Copy(io.Discard, resp.Body)
	if errcoppy != nil {
		return err
	}
	return nil
}
