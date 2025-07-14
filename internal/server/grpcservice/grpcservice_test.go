package grpcmetrics

import (
	"context"
	"io"
	"log"
	"net"
	"testing"

	pb "github.com/4aleksei/metricscum/internal/common/grpcmetrics/proto"
	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
	"github.com/stretchr/testify/assert"

	"github.com/4aleksei/metricscum/internal/common/repository/memstorage"
	"github.com/4aleksei/metricscum/internal/common/utils"
	"github.com/4aleksei/metricscum/internal/server/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type ValueName struct {
	name  string
	value valuemetric.ValueMetric
}

func (v *ValueName) getMetric() *pb.Metric {
	return &pb.Metric{
		Name:    v.name,
		Counter: utils.Setint64(v.value.ValueInt()),
		Gauge:   utils.Setfloat64(v.value.ValueFloat()),
		Type:    pb.Metric_Type(v.value.GetKind()),
	}
}

func (v *ValueName) getValue(b *pb.Metric) error {
	v.name = b.Name
	k, _ := valuemetric.GetKindInt(int(b.GetType()))
	delta := b.GetCounter()
	value := b.GetGauge()
	val, err := valuemetric.ConvertToValueMetricInt(k, &delta, &value)
	if err != nil {
		return err
	}
	v.value = *val
	return nil
}

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func initNew() {
	lis = bufconn.Listen(bufSize)
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestServerSingle(t *testing.T) {
	initNew()
	grpcServer := grpc.NewServer()
	store := service.NewHandlerStore(memstorage.NewStore())

	pb.RegisterStreamMultiServiceServer(grpcServer, StreamMultiService{store: store})

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()
	defer grpcServer.Stop() // Ensure server is stopped after test

	// устанавливаем соединение с сервером
	conn, err := grpc.NewClient("passthrough:///bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewStreamMultiServiceClient(conn)

	tests := []struct {
		name string
		oper string
		val  ValueName
		want ValueName
	}{
		{name: "Test1 set 100", oper: "Update",
			val:  ValueName{name: "testCounter1", value: *valuemetric.ConvertToIntValueMetric(100)},
			want: ValueName{name: "testCounter1", value: *valuemetric.ConvertToIntValueMetric(100)}},
		{name: "Test2 add 100", oper: "Update",
			val:  ValueName{name: "testCounter1", value: *valuemetric.ConvertToIntValueMetric(100)},
			want: ValueName{name: "testCounter1", value: *valuemetric.ConvertToIntValueMetric(200)}},

		{name: "Test3 get 200", oper: "Get",
			val:  ValueName{name: "testCounter1", value: *valuemetric.ConvertToIntValueMetric(0)},
			want: ValueName{name: "testCounter1", value: *valuemetric.ConvertToIntValueMetric(200)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var val *pb.Response
			var resp ValueName
			var err error
			switch tt.oper {
			case "Update":
				val, err = client.UpdateRequest(context.Background(), &pb.Request{Value: tt.val.getMetric()})
			case "Get":
				val, err = client.GetMetric(context.Background(), &pb.Request{Value: tt.val.getMetric()})
			default:
				log.Fatal("error operation")
			}

			if err != nil {
				if e, ok := status.FromError(err); ok {
					switch e.Code() {
					case codes.NotFound, codes.DeadlineExceeded:
						log.Println(e.Message())
					default:
						log.Println(e.Code(), e.Message())
					}
				} else {
					log.Printf("не получилось распарсить ошибку %v", err)
				}
			}
			resp.getValue(val.Value)
			assert.Equal(t, tt.want, resp)
		})
	}

}

func TestServerMulti(t *testing.T) {
	initNew()
	grpcServer := grpc.NewServer()
	store := service.NewHandlerStore(memstorage.NewStore())

	pb.RegisterStreamMultiServiceServer(grpcServer, StreamMultiService{store: store})

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()
	defer grpcServer.Stop() // Ensure server is stopped after test

	// устанавливаем соединение с сервером
	conn, err := grpc.NewClient("passthrough:///bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewStreamMultiServiceClient(conn)

	sliceTest := []ValueName{
		ValueName{name: "testCounter1", value: *valuemetric.ConvertToIntValueMetric(100)},
		ValueName{name: "testCounter2", value: *valuemetric.ConvertToIntValueMetric(100)},
	}

	tests := []struct {
		name string
		oper string
		val  []ValueName
		want []ValueName
	}{
		{name: "Test1 set 100", oper: "Update",
			val:  sliceTest,
			want: sliceTest,
		},
		{name: "Test2 get 100", oper: "Get",

			want: sliceTest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.oper {
			case "Update":

				var metrics []*pb.Metric
				for _, val := range tt.val {
					metrics = append(metrics, &pb.Metric{
						Name:    val.name,
						Counter: utils.Setint64(val.value.ValueInt()),
						Gauge:   utils.Setfloat64(val.value.ValueFloat()),
						Type:    pb.Metric_Type(val.value.GetKind()),
					})
				}

				val, err := client.MultiUpdateRequest(context.Background(), &pb.MultiUpdate{Values: metrics})

				if err != nil {
					if e, ok := status.FromError(err); ok {
						switch e.Code() {
						case codes.NotFound, codes.DeadlineExceeded:
							log.Println(e.Message())
						default:
							log.Println(e.Code(), e.Message())
						}
					} else {
						log.Printf("не получилось распарсить ошибку %v", err)
					}
				}
				var respValues []ValueName
				for _, val := range val.GetValues() {
					var v ValueName
					v.getValue(val)
					respValues = append(respValues, v)
				}

				assert.ElementsMatch(t, tt.want, respValues)

			case "Get":
				stream, err := client.GetMetrics(context.Background(), &pb.RequestMetrics{})

				if err != nil {
					if e, ok := status.FromError(err); ok {
						switch e.Code() {
						case codes.NotFound, codes.DeadlineExceeded:
							log.Println(e.Message())
						default:
							log.Println(e.Code(), e.Message())
						}
					} else {
						log.Printf("не получилось распарсить ошибку %v", err)
					}
				}

				var respValues []ValueName
				for {
					resp, err := stream.Recv()
					if err == io.EOF {
						break
					}
					if err != nil {
						log.Fatal(err)
					}
					var v ValueName
					v.getValue(resp)
					respValues = append(respValues, v)
				}
				assert.ElementsMatch(t, tt.want, respValues)
			default:
				log.Fatal("error operation")
			}

		})
	}

}
