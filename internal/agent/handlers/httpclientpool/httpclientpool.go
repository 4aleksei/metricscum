package httpclientpool

import (
	"context"
	"errors"

	"io"
	"net/http"
	"sync"

	"net"
	"time"

	"github.com/4aleksei/metricscum/internal/agent/config"
	"github.com/4aleksei/metricscum/internal/agent/handlers/httpclientpool/job"

	"bytes"
	"compress/gzip"
	"encoding/hex"

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
		WorkerCount int

		//	jobs        chan job.Job
		//	results     chan job.Result
		//	wg          sync.WaitGroup
		cfg     *config.Config
		clients []clientInstance
		// cancels     []context.CancelFunc
	}
	functioExec func(context.Context, *sync.WaitGroup, *http.Client,
		<-chan job.Job, chan<- job.Result, *config.Config)
	clientInstance struct {
		execFn functioExec
		client *http.Client
		cfg    *config.Config
		// cancels context.CancelFunc
	}
)

func NewHandler(cfg *config.Config) *PoolHandler {
	p := new(PoolHandler)
	p.WorkerCount = int(cfg.RateLimit)
	p.clients = make([]clientInstance, p.WorkerCount)
	p.cfg = cfg
	for i := 0; i < p.WorkerCount; i++ {
		p.clients[i] = *newClientInstance(cfg)
	}
	return p
}

func newClientInstance(cfg *config.Config) *clientInstance {
	return &clientInstance{
		execFn: poolOptions(cfg),
		client: newClient(),
		cfg:    cfg,
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

func newClient() *http.Client {
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 2 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 2 * time.Second,
	}
	return &http.Client{
		Transport: netTransport,
	}
}


func workerJSONBatch(ctx context.Context, wg *sync.WaitGroup, client *http.Client,
	jobs <-chan job.Job, results chan<- job.Result, cfg *config.Config) {
	defer wg.Done()
	server := "http://" + cfg.Address + "/updates/"
	for j := range jobs {
		select {
		case <-ctx.Done():
			return
		default:

			err := jsonModelSFunc(ctx, server, client, j.Value, cfg.Key)
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


func workerJSON(ctx context.Context, wg *sync.WaitGroup, client *http.Client,
	jobs <-chan job.Job, results chan<- job.Result, cfg *config.Config) {
	defer wg.Done()
	server := "http://" + cfg.Address + "/update/"
	for j := range jobs {
		select {
		case <-ctx.Done():
			return
		default:

			err := jsonModelFunc(ctx, server, client, &j.Value[0], cfg.Key)
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

func workerPlain(ctx context.Context, wg *sync.WaitGroup, client *http.Client,
	jobs <-chan job.Job, results chan<- job.Result, cfg *config.Config) {
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


func (p *PoolHandler) StartPool(ctx context.Context, jobs chan job.Job, results chan job.Result, wg *sync.WaitGroup) {
	for i := 0; i < int(p.WorkerCount); i++ {
		wg.Add(1)
		go p.clients[i].execFn(ctx, wg, p.clients[i].client, jobs, results, p.cfg)
	}
}

/*func (p *PoolHandler) GetResult(ctx context.Context) (job.Result, error) {

	select {
	case <-ctx.Done():
		return job.Result{}, ctx.Err()
	case res, ok := <-p.results:
		if !ok {
			return job.Result{}, ErrChanClosed
		}
		return res, nil
	}

}*/

func (p *PoolHandler) Start(ctx context.Context) error {
	/*	for i := 0; i < int(p.workerCount); i++ {
		p.wg.Add(1)
		ctxCancel, cancel := context.WithCancel(context.Background())
		p.cancels = append(p.cancels, cancel)
		go p.execFn(ctxCancel, newClient(), &p.wg, p.jobs, p.results, p.cfg)*/
	return nil
}

func (p *PoolHandler) Stop(ctx context.Context) error {

	/*	for i := 0; i < len(p.cancels); i++ {
			p.cancels[i]()
		}
		p.wg.Wait()
		close(p.jobs)
		close(p.results)*/
	return nil
}

func plainTxtFunc(ctx context.Context, client *http.Client, server, data string) error {
	err := utils.RetryAction(ctx, utils.RetryTimes(), func(ctx context.Context) error {
		return newPPostReq(ctx, client, server+data, http.NoBody)
	})
	if err != nil {
		return err
	}
	return nil
}

func jsonModelFunc(ctx context.Context, server string, client *http.Client, data *models.Metrics, cfgkey string) error {
	var requestBody bytes.Buffer

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
	err := data.JSONEncodeBytes(twr)
	if err != nil {
		return err
	}
	gz.Close()

	if hmac != nil {
		key = hex.EncodeToString(hmac.GetSig())
	}

	err = utils.RetryAction(ctx, utils.RetryTimes(), func(ctx context.Context) error {
		return newJPostReq(ctx, client, server, &requestBody, key)
	})
	if err != nil {
		return err
	}
	return nil
}
func jsonModelSFunc(ctx context.Context, server string, client *http.Client, data []models.Metrics, cfgkey string) error {
	var requestBody bytes.Buffer
	var key string
	var twr io.Writer
	gz := gzip.NewWriter(&requestBody)
	var hmac *hmacsha256.HmacWriter

	if cfgkey != "" {
		hmac = hmacsha256.NewWriter(twr, []byte(cfgkey))
		twr = hmac
	} else {
		twr = gz
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
		return newJPostReq(ctx, client, server, &requestBody, key)
	})
	if err != nil {
		return err
	}
	return nil
}

func newPPostReq(ctx context.Context, client *http.Client, server string, requestBody io.Reader) error {
	req, err := http.NewRequestWithContext(ctx, "POST", server, requestBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", textPlainContent)
	resp, err := client.Do(req)
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

func newJPostReq(ctx context.Context, client *http.Client, server string, requestBody io.Reader, key string) error {
	req, err := http.NewRequestWithContext(ctx, "POST", server, requestBody)
	if err != nil {
		return err
	}

	if key != "" {
		req.Header.Set("HashSHA256", key)
	}

	req.Header.Set("Accept-Encoding", gzipContent)
	req.Header.Set("Content-Encoding", gzipContent)
	req.Header.Set("Content-Type", applicationJSONContent)
	req.Header.Set("Accept", applicationJSONContent)
	resp, err := client.Do(req)
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
