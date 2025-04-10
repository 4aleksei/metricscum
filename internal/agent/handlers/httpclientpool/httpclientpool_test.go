package httpclientpool

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/4aleksei/metricscum/internal/agent/config"
	"github.com/4aleksei/metricscum/internal/agent/handlers/httpclientpool/job"
	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/stretchr/testify/assert"
)

func Test_NewPool(t *testing.T) {
	cfg := &config.Config{
		Address:     "127.0.0.1:8081",
		RateLimit:   1,
		ContentJSON: false,
	}

	t.Run("Test NewPool", func(t *testing.T) {
		p := NewHandler(cfg)
		assert.NotNil(t, p)
	})
}

func Test_Plain(t *testing.T) {
	cfg := &config.Config{
		Address:     "127.0.0.1:8081",
		RateLimit:   1,
		ContentJSON: false,
	}
	assert := assert.New(t)
	l, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		log.Fatal(err)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		_, _ = rw.Write([]byte(`OK`))
	}))
	server.Listener.Close()
	server.Listener = l

	server.Start()
	defer server.Close()

	p := new(PoolHandler)
	p.WorkerCount = int(cfg.RateLimit)
	p.clients = make([]clientInstance, p.WorkerCount)
	p.cfg = cfg
	p.clients[0] = clientInstance{
		execFn: poolOptions(cfg),
		client: server.Client(),
		cfg:    cfg,
	}

	wg := &sync.WaitGroup{}
	jobs := make(chan job.Job, p.WorkerCount*2)
	results := make(chan job.Result, p.WorkerCount*2)
	val := make([]models.Metrics, 0)
	var valint int64 = 100
	val = append(val, models.Metrics{ID: "TEst", MType: "counter", Delta: &valint})

	p.StartPool(context.Background(), jobs, results, wg)
	id := job.JobID(1)
	jobs <- job.Job{ID: id, Value: val}
	res := <-results

	assert.Equal(nil, res.Err, "No error needed")

	close(jobs)
	wg.Wait()
	close(results)
}

func Test_Json(t *testing.T) {
	cfg := &config.Config{
		Address:     "127.0.0.1:8081",
		RateLimit:   1,
		ContentJSON: true,
	}
	assert := assert.New(t)
	l, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		log.Fatal(err)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		_, _ = rw.Write([]byte(`OK`))
	}))
	server.Listener.Close()
	server.Listener = l

	// Start the server.
	server.Start()
	defer server.Close()

	p := new(PoolHandler)
	p.WorkerCount = int(cfg.RateLimit)
	p.clients = make([]clientInstance, p.WorkerCount)
	p.cfg = cfg
	p.clients[0] = clientInstance{
		execFn: poolOptions(cfg),
		client: server.Client(),
		cfg:    cfg,
	}

	wg := &sync.WaitGroup{}
	jobs := make(chan job.Job, p.WorkerCount*2)
	results := make(chan job.Result, p.WorkerCount*2)
	val := make([]models.Metrics, 0)
	var valint int64 = 100
	val = append(val, models.Metrics{ID: "TEst", MType: "counter", Delta: &valint})

	p.StartPool(context.Background(), jobs, results, wg)
	id := job.JobID(1)
	jobs <- job.Job{ID: id, Value: val}
	res := <-results

	assert.Equal(nil, res.Err, "No error needed")

	close(jobs)
	wg.Wait()
	close(results)
}

func Test_JsonBatch(t *testing.T) {
	cfg := &config.Config{
		Address:      "127.0.0.1:8081",
		RateLimit:    1,
		ContentJSON:  true,
		ContentBatch: 1,
	}
	assert := assert.New(t)
	l, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		log.Fatal(err)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		_, _ = rw.Write([]byte(`OK`))
	}))
	server.Listener.Close()
	server.Listener = l

	// Start the server.
	server.Start()
	defer server.Close()

	p := new(PoolHandler)
	p.WorkerCount = int(cfg.RateLimit)
	p.clients = make([]clientInstance, p.WorkerCount)
	p.cfg = cfg
	p.clients[0] = clientInstance{
		execFn: poolOptions(cfg),
		client: server.Client(),
		cfg:    cfg,
	}

	wg := &sync.WaitGroup{}
	jobs := make(chan job.Job, p.WorkerCount*2)
	results := make(chan job.Result, p.WorkerCount*2)
	val := make([]models.Metrics, 0)
	var valint int64 = 100
	val = append(val, models.Metrics{ID: "TEst", MType: "counter", Delta: &valint})

	p.StartPool(context.Background(), jobs, results, wg)
	id := job.JobID(1)
	jobs <- job.Job{ID: id, Value: val}
	res := <-results

	assert.Equal(nil, res.Err, "No error needed")

	close(jobs)
	wg.Wait()
	close(results)
}
