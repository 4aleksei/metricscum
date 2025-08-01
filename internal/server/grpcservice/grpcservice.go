// Package grpcmetrics - gprc service
package grpcmetrics

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/netip"

	"google.golang.org/grpc/credentials"

	pb "github.com/4aleksei/metricscum/internal/common/grpcmetrics/proto"
	"github.com/4aleksei/metricscum/internal/common/models"
	"github.com/4aleksei/metricscum/internal/common/repository/valuemetric"
	"github.com/4aleksei/metricscum/internal/common/utils"
	"github.com/4aleksei/metricscum/internal/server/config"
	"github.com/4aleksei/metricscum/internal/server/service"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type StreamMultiService struct {
	pb.UnimplementedStreamMultiServiceServer
	store *service.HandlerStore
	srv   *grpc.Server
	l     *zap.Logger
	cfg   *config.Config
}

func (s StreamMultiService) UpdateRequest(ctx context.Context, in *pb.Request) (*pb.Response, error) {
	var response pb.Response
	var valModel models.Metrics
	valModel.ConvertToModel(in.GetValue())
	val, err := s.store.SetValueModel(ctx, valModel)
	if err != nil {
		return nil, status.Errorf(codes.Internal, `%s`, err.Error())
	} else {
		k, _ := valuemetric.GetKind(val.MType)
		response.Value = &pb.Metric{
			Name:    val.ID,
			Counter: utils.Setint64(val.Delta),
			Gauge:   utils.Setfloat64(val.Value),
			Type:    pb.Metric_Type(k),
		}
	}
	return &response, nil
}

func (s StreamMultiService) GetMetric(ctx context.Context, in *pb.Request) (*pb.Response, error) {
	var response pb.Response
	var valModel models.Metrics
	valModel.ConvertToModel(in.GetValue())
	val, err := s.store.GetValueModel(ctx, valModel)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, `%s`, err.Error())
	} else {
		k, _ := valuemetric.GetKind(val.MType)
		response.Value = &pb.Metric{
			Name:    val.ID,
			Counter: utils.Setint64(val.Delta),
			Gauge:   utils.Setfloat64(val.Value),
			Type:    pb.Metric_Type(k),
		}
	}
	return &response, nil
}

func (s StreamMultiService) MultiUpdateRequest(ctx context.Context, in *pb.MultiUpdate) (*pb.MultiResponse, error) {
	var response pb.MultiResponse
	var valModels []models.Metrics

	for _, val := range in.GetValues() {
		var valMod models.Metrics
		valMod.ConvertToModel(val)
		valModels = append(valModels, valMod)
	}
	resp, err := s.store.SetValueSModel(ctx, valModels)
	if err != nil {
		return nil, status.Errorf(codes.Internal, `%s`, err.Error())
	}
	var metrics []*pb.Metric

	for _, val := range resp {
		k, _ := valuemetric.GetKind(val.MType)
		metrics = append(metrics, &pb.Metric{
			Name:    val.ID,
			Counter: utils.Setint64(val.Delta),
			Gauge:   utils.Setfloat64(val.Value),
			Type:    pb.Metric_Type(k),
		})
	}
	response.Values = metrics
	return &response, nil
}

func (s StreamMultiService) GetMetrics(in *pb.RequestMetrics, srv pb.StreamMultiService_GetMetricsServer) error {
	err := s.store.GetAllStoreValue(srv.Context(), func(key string, val valuemetric.ValueMetric) error {
		resp := pb.Metric{
			Name:    key,
			Counter: utils.Setint64(val.ValueInt()),
			Gauge:   utils.Setfloat64(val.ValueFloat()),
			Type:    pb.Metric_Type(val.GetKind()),
		}
		if err := srv.Send(&resp); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func getTls(cfg *config.Config) (*tls.Config, error) {
	serverCert, err := tls.LoadX509KeyPair(cfg.PrivateCertFile, cfg.PrivateKeyFile)
	if err != nil {
		return nil, err
	}

	// Create TLS configuration
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.NoClientCert, // Or tls.RequireAndVerifyClientCert for mTLS
	}

	// Create gRPC server credentials
	return config, nil
}

func NewgPRC(s *service.HandlerStore, cfg *config.Config, l *zap.Logger) (*StreamMultiService, error) {
	listen, err := net.Listen("tcp", cfg.Grcp)
	if err != nil {
		l.Debug("gRCP Listen Error: ", zap.Error(err))
		return nil, err
	}

	opts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}
	var trustedPeers []net.IPNet
	if cfg.Cidr != "" {
		_, trustedCidr, err := net.ParseCIDR(cfg.Cidr)
		if err != nil {
			return nil, err
		}
		trustedPeers = []net.IPNet{
			*trustedCidr,
		}
	}

	optsMy := []Option{
		WithTrustedCidr(trustedPeers),
	}

	var grpcServer *grpc.Server

	configTls, err := getTls(cfg)
	if err != nil {
		return nil, err
	}
	if configTls != nil {
		tlsCredential := credentials.NewTLS(configTls)
		grpcServer = grpc.NewServer(grpc.Creds(tlsCredential),
			grpc.ChainUnaryInterceptor(
				logging.UnaryServerInterceptor(InterceptorLogger(l), opts...),
				UnaryServerBlock(optsMy...),
			),
			grpc.ChainStreamInterceptor(
				logging.StreamServerInterceptor(InterceptorLogger(l), opts...),
				StreamServerBlock(optsMy...),
			),
		)
	} else {
		grpcServer = grpc.NewServer(grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(InterceptorLogger(l), opts...),
			UnaryServerBlock(optsMy...),
		),
			grpc.ChainStreamInterceptor(
				logging.StreamServerInterceptor(InterceptorLogger(l), opts...),
				StreamServerBlock(optsMy...),
			),
		)
	}

	serv := &StreamMultiService{store: s,
		srv: grpcServer,
		l:   l,
		cfg: cfg}

	pb.RegisterStreamMultiServiceServer(grpcServer, serv)

	go func() {
		if err := grpcServer.Serve(listen); err != nil {
			l.Debug("gRCP Server Start Error: ", zap.Error(err))
		}
	}()
	return serv, nil
}

func (s StreamMultiService) StopServ() {
	s.srv.GracefulStop()
	s.srv.Stop()
}

func InterceptorLogger(l *zap.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		f := make([]zap.Field, 0, len(fields)/2)

		for i := 0; i < len(fields); i += 2 {
			key := fields[i]
			value := fields[i+1]

			switch v := value.(type) {
			case string:
				f = append(f, zap.String(key.(string), v))
			case int:
				f = append(f, zap.Int(key.(string), v))
			case bool:
				f = append(f, zap.Bool(key.(string), v))
			default:
				f = append(f, zap.Any(key.(string), v))
			}
		}

		logger := l.WithOptions(zap.AddCallerSkip(1)).With(f...)

		switch lvl {
		case logging.LevelDebug:
			logger.Debug(msg)
		case logging.LevelInfo:
			logger.Info(msg)
		case logging.LevelWarn:
			logger.Warn(msg)
		case logging.LevelError:
			logger.Error(msg)
		default:

		}
	})
}

var (
	ErrNoTrust = errors.New("reject")
)

type options struct {
	trustedCidr []net.IPNet
}

type Option func(*options)

func evaluateOpts(opts []Option) *options {
	optCopy := &options{}
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

func WithTrustedCidr(peers []net.IPNet) Option {
	return func(o *options) {
		o.trustedCidr = peers
	}
}

func UnaryServerBlock(opts ...Option) grpc.UnaryServerInterceptor {
	o := evaluateOpts(opts)
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if len(o.trustedCidr) == 0 {
			return handler(ctx, req)
		}
		realIP := getPeerAddr(ctx)
		ip := net.ParseIP(realIP)
		if ip == nil {
			return nil, ErrNoTrust
		}
		if o.trustedCidr[0].Contains(ip) {
			return handler(ctx, req)
		}
		return nil, ErrNoTrust
	}
}

func getPeerAddr(ctx context.Context) string {
	var realIP string
	vals := metadata.ValueFromIncomingContext(ctx, "X-Real-IP")
	if len(vals) != 0 {
		realIP = vals[0]
	} else {
		pr := remotePeer(ctx)
		addrPort, _ := netip.ParseAddrPort(pr.String())
		realIP = addrPort.Addr().String()
	}
	return realIP
}

func StreamServerBlock(opts ...Option) grpc.StreamServerInterceptor {
	o := evaluateOpts(opts)
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if len(o.trustedCidr) == 0 {
			return handler(srv, stream)
		}
		realIP := getPeerAddr(stream.Context())
		ip := net.ParseIP(realIP)

		if o.trustedCidr[0].Contains(ip) {
			return handler(srv, stream)
		}
		return ErrNoTrust
	}
}

func remotePeer(ctx context.Context) net.Addr {
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return nil
	}
	return pr.Addr
}
