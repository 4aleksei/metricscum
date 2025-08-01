// Package grpcclient service client
package grpcclient

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	pb "github.com/4aleksei/metricscum/internal/common/grpcmetrics/proto"
	"github.com/4aleksei/metricscum/internal/common/models"

	"crypto/rsa"
	"crypto/x509"

	"github.com/4aleksei/metricscum/internal/agent/config"
	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
	"github.com/4aleksei/metricscum/internal/common/utils"
	"google.golang.org/grpc"

	"github.com/4aleksei/metricscum/internal/common/job"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
)

type (
	GRPCPool struct {
		cfg         *config.Config
		WorkerCount int
		clients     []clientInstance
	}
	functioExec func(context.Context, *clientInstance, *sync.WaitGroup, <-chan job.Job, chan<- job.Result)

	clientInstance struct {
		execFn    functioExec
		client    *agentClient
		cfg       *config.Config
		publicKey *rsa.PublicKey
	}
	agentClient struct {
		client     pb.StreamMultiServiceClient
		connection *grpc.ClientConn
		localAddr  string
	}
)

func newClientInstance(cfg *config.Config, p *rsa.PublicKey) *clientInstance {
	return &clientInstance{
		execFn:    poolOptions(cfg),
		client:    newClient(cfg),
		cfg:       cfg,
		publicKey: p,
	}
}

func poolOptions(cfg *config.Config) functioExec {
	if cfg.ContentJSON {
		if cfg.ContentBatch > 0 {
			return workerBatch
		} else {
			return workerSingle
		}
	} else {
		return workerSingle
	}
}

func loadCert(cfg *config.Config) (grpc.DialOption, error) {

	if cfg.CertKeyFile != "" {
		caCert, err := os.ReadFile(cfg.CertKeyFile)
		if err != nil {
			return grpc.WithTransportCredentials(insecure.NewCredentials()), err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		// Create TLS configuration
		tlsCredentials := credentials.NewClientTLSFromCert(caCertPool, "") // "localhost" must match server certificate's CN or SAN

		return grpc.WithTransportCredentials(tlsCredentials), nil
	}

	return grpc.WithTransportCredentials(insecure.NewCredentials()), nil
}

func newClient(cfg *config.Config) *agentClient {
	agclient := &agentClient{}
	myDialer := net.Dialer{Timeout: 30 * time.Second,
		KeepAlive: 30 * time.Second}

	dialOpt, err := loadCert(cfg)
	if err != nil {
		fmt.Println("Error load cert ", err)
	}
	conn, err := grpc.NewClient(cfg.Address, dialOpt,
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			conn, err := myDialer.DialContext(ctx, "tcp", addr)
			if err == nil {
				agclient.localAddr = conn.LocalAddr().(*net.TCPAddr).IP.String()
			}
			return conn, err
		}))

	if err != nil {
		fmt.Println("Error grpc.NewClient ", err)
	}

	c := pb.NewStreamMultiServiceClient(conn)
	agclient.client = c
	agclient.connection = conn
	return agclient
}

func NewgRPC(cfg *config.Config) *GRPCPool {
	p := &GRPCPool{
		WorkerCount: int(cfg.RateLimit),
		cfg:         cfg,
		clients:     make([]clientInstance, int(cfg.RateLimit)),
	}

	for i := 0; i < p.WorkerCount; i++ {
		p.clients[i] = *newClientInstance(cfg, nil)
	}
	return p
}

func (p *GRPCPool) GracefulStop() {
	for _, v := range p.clients {
		v.client.connection.Close()
	}
}

func (p *GRPCPool) StartPool(ctx context.Context, jobs chan job.Job, results chan job.Result, wg *sync.WaitGroup) {
	for i := 0; i < p.WorkerCount; i++ {
		wg.Add(1)
		go p.clients[i].execFn(ctx, &p.clients[i], wg, jobs, results)
	}
}

func workerBatch(ctx context.Context, c *clientInstance, wg *sync.WaitGroup, jobs <-chan job.Job, results chan<- job.Result) {
	defer wg.Done()
	for j := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
			err := sendBatch(ctx, c.client, j.Value)
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

func workerSingle(ctx context.Context, c *clientInstance, wg *sync.WaitGroup, jobs <-chan job.Job, results chan<- job.Result) {
	defer wg.Done()
	for j := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
			err := sendSingle(ctx, c.client, &j.Value[0])
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

func sendSingle(ctx context.Context, client *agentClient, data *models.Metrics) error {
	k, _ := valuemetric.GetKind(data.MType)
	md := metadata.New(map[string]string{"X-Real-IP": client.localAddr})
	ctxReq := metadata.NewOutgoingContext(ctx, md)
	_, err := client.client.UpdateRequest(ctxReq, &pb.Request{Value: &pb.Metric{
		Name:    data.ID,
		Counter: utils.Setint64(data.Delta),
		Gauge:   utils.Setfloat64(data.Value),
		Type:    pb.Metric_Type(k),
	}},
		grpc.UseCompressor(gzip.Name))

	if err != nil {
		return err
	}
	return nil
}

func sendBatch(ctx context.Context, client *agentClient, data []models.Metrics) error {
	var metrics []*pb.Metric
	for _, val := range data {
		k, _ := valuemetric.GetKind(val.MType)
		metrics = append(metrics, &pb.Metric{
			Name:    val.ID,
			Counter: utils.Setint64(val.Delta),
			Gauge:   utils.Setfloat64(val.Value),
			Type:    pb.Metric_Type(k),
		})
	}
	md := metadata.New(map[string]string{"X-Real-IP": client.localAddr})
	ctxReq := metadata.NewOutgoingContext(ctx, md)
	_, err := client.client.MultiUpdateRequest(ctxReq, &pb.MultiUpdate{Values: metrics}, grpc.UseCompressor(gzip.Name))
	if err != nil {
		return err
	}
	return nil
}
