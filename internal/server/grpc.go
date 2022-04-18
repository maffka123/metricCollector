package server

import (
	"context"

	"github.com/maffka123/metricCollector/internal/handlers"
	"github.com/maffka123/metricCollector/internal/models"
	pb "github.com/maffka123/metricCollector/internal/proto"
	"github.com/maffka123/metricCollector/internal/storage"
	"go.uber.org/zap"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net"
)

type MetricsServer struct {
	// нужно встраивать тип pb.Unimplemented<TypeName>
	// для совместимости с будущими версиями
	pb.UnimplementedMetricsServer
	db  storage.Repositories
	log *zap.Logger
}

func (s *MetricsServer) AddMetricGauge(ctx context.Context, in *pb.Metric) (*pb.StdResponse, error) {
	var response pb.StdResponse

	s.db.InsertGouge(in.Id, float64(in.Value))
	s.log.Info("Added new gauge")

	return &response, nil
}

func (s *MetricsServer) AddMetricCounter(ctx context.Context, in *pb.Metric) (*pb.StdResponse, error) {
	var response pb.StdResponse

	s.db.InsertCounter(in.Id, in.Delta)
	s.log.Info("Added new counter")

	return &response, nil
}

func (s *MetricsServer) GetMetricValue(ctx context.Context, in *pb.Metric) (*pb.GetMetricValueResponse, error) {
	var response pb.GetMetricValueResponse
	if in.MType == "counter" && s.db.NameInCounter(in.Id) {
		response.Delta = s.db.ValueFromCounter(in.Id)
	} else if in.MType == "gauge" && s.db.NameInGouge(in.Id) {
		response.Value = s.db.ValueFromGouge(in.Id)
	}
	s.log.Info("Found metric value")

	return &response, nil
}

func (s *MetricsServer) GetMetricNames(ctx context.Context, in *pb.Empty) (*pb.GetMetricNamesResponse, error) {
	var response pb.GetMetricNamesResponse

	g, c := s.db.SelectAll()

	response.Names = append(g, c...)
	s.log.Info("Recieved all metrics names")

	return &response, nil
}

func (s *MetricsServer) UpdateMetric(ctx context.Context, in *pb.Metric) (*pb.StdResponse, error) {
	var response pb.StdResponse
	if in.MType == "counter" && s.db.NameInCounter(in.Id) {
		s.db.InsertCounter(in.Id, in.Delta)
	} else if in.MType == "gauge" && s.db.NameInGouge(in.Id) {
		s.db.InsertGouge(in.Id, in.Value)
	}
	s.log.Info("Updated metric")

	return &response, nil
}
func (s *MetricsServer) GetMetric(ctx context.Context, in *pb.Metric) (*pb.Metric, error) {
	var response pb.Metric

	if in.MType == "counter" && s.db.NameInCounter(in.Id) {
		response.Delta = s.db.ValueFromCounter(in.Id)
	} else if in.MType == "gauge" && s.db.NameInGouge(in.Id) {
		response.Value = s.db.ValueFromGouge(in.Id)
	}
	s.log.Info("Found metric value")
	return &response, nil
}
func (s *MetricsServer) UpdateMetrics(ctx context.Context, in *pb.UpdateMetricsRequest) (*pb.StdResponse, error) {
	var response pb.StdResponse
	var ml []models.Metrics
	for _, m := range in.Metrics {
		a := models.Metrics{ID: m.Id, MType: m.MType, Delta: &m.Delta, Value: &m.Value, Hash: m.Hash}
		ml = append(ml, a)
	}
	s.db.BatchInsert(ml)
	s.log.Info("Inserted new metrics into db")

	return &response, nil
}

func NewGRPCServer(endpoint string, db storage.Repositories, subnet string, log *zap.Logger) (*grpc.Server, net.Listener) {
	listen, err := net.Listen("tcp", endpoint)
	if err != nil {
		log.Error("starting tcp tunnel failed", zap.Error(err))
	}
	// создаём gRPC-сервер без зарегистрированной службы
	si := subnetInterceptor{subnet: subnet}
	s := grpc.NewServer(grpc.UnaryInterceptor(si.serverInterceptor()))
	// регистрируем сервис
	pb.RegisterMetricsServer(s, &MetricsServer{db: db, log: log})

	return s, listen
	// получаем запрос gRpc
}

type subnetInterceptor struct {
	subnet string
}

func (si *subnetInterceptor) serverInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "Retrieving metadata is failed")
		}
		authHeader, ok := md["x-real-ip"]

		if si.subnet != "" {
			b, err := handlers.IfIPinCIDR(si.subnet, authHeader[0])
			if err != nil {
				return nil, status.Errorf(codes.PermissionDenied, "IP address could not be parsed: %s", err)

			}

			if !*b {
				return nil, status.Errorf(codes.PermissionDenied, "IP address is not inside relible network")
			}
		}

		return handler(ctx, req)
	}
}
