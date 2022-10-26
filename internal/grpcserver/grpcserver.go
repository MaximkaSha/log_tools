package grpcserver

import (
	"context"
	"net"
	"strings"

	"github.com/MaximkaSha/log_tools/internal/handlers"
	"github.com/MaximkaSha/log_tools/internal/models"
	pb "github.com/MaximkaSha/log_tools/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type MetricsServer struct {
	pb.UnimplementedMetricsServer

	handl handlers.Handlers
}

func NewMetricServer(handl handlers.Handlers) *MetricsServer {
	return &MetricsServer{
		handl: handl,
	}
}

func (m MetricsServer) AddMetric(ctx context.Context, in *pb.AddMetricRequest) (*pb.AddMetricResponse, error) {

	data := models.NewMetric(
		in.Metric.Id,
		strings.ToLower(in.Metric.Mtype.String()),
		&in.GetMetric().Delta,
		&in.GetMetric().Value,
		in.Metric.Hash)
	if m.handl.CryptoService.IsEnable {
		if m.handl.CryptoService.CheckHash(data) {
			response := pb.AddMetricResponse{}
			return &response, status.Errorf(codes.DataLoss, `Check sign error`)
		}
	}
	err := m.handl.Repo.InsertMetric(ctx, data)
	if err != nil {
		response := pb.AddMetricResponse{}
		return &response, status.Errorf(codes.InvalidArgument, "Invalid Metric")
	}
	response := pb.AddMetricResponse{}
	return &response, nil
}

func (m MetricsServer) AddMetrics(ctx context.Context, in *pb.AddMetricsRequest) (*pb.AddMetricsResponse, error) {
	allData := []models.Metrics{}
	// Вот тут не очень конечно. Конвертируем из json в protobuf, потом опять из protobuf в json.
	// Надо наверное сделать две раздельные хранилки в агенте для json и protobuff.

	for i := range in.Metrics {
		data := models.NewMetric(
			in.Metrics[i].Id,
			strings.ToLower(in.Metrics[i].Mtype.String()),
			&in.Metrics[i].Delta,
			&in.Metrics[i].Value,
			in.Metrics[i].Hash)
		// кроме того я так и не понял как получить доступ к данным из интерцептора
		if m.handl.CryptoService.IsEnable {
			if m.handl.CryptoService.CheckHash(data) {
				response := pb.AddMetricsResponse{}
				return &response, status.Errorf(codes.InvalidArgument, `Check sign error`)
			}
		}
		allData = append(allData, data)
	}
	m.handl.Repo.BatchInsert(ctx, allData)
	response := pb.AddMetricsResponse{}
	return &response, nil
}

func (m MetricsServer) UnaryCheckIPInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "Retrieving metadata is failed")
	}
	data, ok := md["x-real-ip"]
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "missing IP")
	}
	ip := net.ParseIP(data[0])
	if ip == nil || !m.handl.TrustedSubnet.Contains(ip) {
		return nil, status.Errorf(codes.Unauthenticated, "not in trusted subnet")
	}

	h, err := handler(ctx, req)
	return h, err

}
