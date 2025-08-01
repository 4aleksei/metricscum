package grpcclient

import (
	"context"
	"log"
	"net"
	"sync"
	"testing"

	"github.com/4aleksei/metricscum/internal/agent/config"
	pb "github.com/4aleksei/metricscum/internal/common/grpcmetrics/proto"
	"github.com/4aleksei/metricscum/internal/common/job"
	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type MockStreamMultiService struct {
	pb.UnimplementedStreamMultiServiceServer
}

func (s MockStreamMultiService) UpdateRequest(ctx context.Context, in *pb.Request) (*pb.Response, error) {
	var response pb.Response
	response.Value = in.GetValue()
	return &response, nil
}

func (s MockStreamMultiService) MultiUpdateRequest(ctx context.Context, in *pb.MultiUpdate) (*pb.MultiResponse, error) {
	var response pb.MultiResponse
	response.Values = in.GetValues()
	return &response, nil
}

func (s MockStreamMultiService) GetMetric(ctx context.Context, in *pb.Request) (*pb.Response, error) {
	var response pb.Response
	response.Value = in.GetValue()
	return &response, nil
}

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func initNew() {
	lis = bufconn.Listen(bufSize)
}
func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func Test_GRCP(t *testing.T) {
	cfg := &config.Config{
		RateLimit:   1,
		ContentJSON: true,
	}
	initNew()
	grpcServer := grpc.NewServer()
	pb.RegisterStreamMultiServiceServer(grpcServer, MockStreamMultiService{})

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()
	defer grpcServer.Stop() // Ensure server is stopped after test

	p := new(GRPCPool)
	p.WorkerCount = int(cfg.RateLimit)
	p.clients = make([]clientInstance, p.WorkerCount)
	p.cfg = cfg

	conn, err := grpc.NewClient("passthrough:///bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewStreamMultiServiceClient(conn)

	p.clients[0] = clientInstance{
		execFn: poolOptions(cfg),
		client: &agentClient{client: client, connection: conn},
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

	assert.Equal(t, nil, res.Err, "No error needed")

	close(jobs)
	wg.Wait()
	close(results)
}
