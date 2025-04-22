package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	pvz_v1 "pvz-service/api/proto_v1"
	"pvz-service/internal/service"
)

type PVZServer struct {
	pvz_v1.UnimplementedPVZServiceServer
	service service.PVZServiceInterface
}

func NewPVZServer(svc service.PVZServiceInterface) *PVZServer {
	return &PVZServer{service: svc}
}

func (s *PVZServer) Register(grpcServer *grpc.Server) {
	pvz_v1.RegisterPVZServiceServer(grpcServer, s)
}

func (s *PVZServer) GetPVZList(ctx context.Context, req *pvz_v1.GetPVZListRequest) (*pvz_v1.GetPVZListResponse, error) {

	// Получаем данные из основного сервиса
	pvzInfos, err := s.service.GetPVZs(ctx)
	if err != nil {
		return nil, err
	}

	// Конвертируем в gRPC формат
	var pvzs []*pvz_v1.PVZ
	for _, info := range pvzInfos {
		pvzs = append(pvzs, &pvz_v1.PVZ{
			Id:               info.ID.String(),
			RegistrationDate: timestamppb.New(info.RegistrationDate),
			City:             info.CityName,
		})
	}

	return &pvz_v1.GetPVZListResponse{Pvzs: pvzs}, nil
}
