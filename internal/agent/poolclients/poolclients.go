// Package poolclients - Facade for clients
package poolclients

import (
	"context"

	"sync"

	"github.com/4aleksei/metricscum/internal/agent/config"
	"github.com/4aleksei/metricscum/internal/agent/grpcclient"
	"github.com/4aleksei/metricscum/internal/agent/handlers/httpclientpool"
	"github.com/4aleksei/metricscum/internal/common/job"
)

type PoolClientI interface {
	StartPool(ctx context.Context, jobs chan job.Job, results chan job.Result, wg *sync.WaitGroup)
	GracefulStop()
}

type PoolClient struct {
	pool        PoolClientI
	WorkerCount int
}

func NewPoolClient(cfg *config.Config) *PoolClient {
	if cfg.Grpc {
		return &PoolClient{
			WorkerCount: int(cfg.RateLimit),
			pool:        grpcclient.NewgRPC(cfg),
		}
	} else {
		return &PoolClient{
			WorkerCount: int(cfg.RateLimit),
			pool:        httpclientpool.NewHandler(cfg),
		}
	}
}

func (po *PoolClient) StartPool(ctx context.Context, jobs chan job.Job, results chan job.Result, wg *sync.WaitGroup) {
	po.pool.StartPool(ctx, jobs, results, wg)
}

func (po *PoolClient) GracefulStop() {
	po.pool.GracefulStop()
}
