package app

import (
	"fmt"
	"log/slog"
	"net"
	"pvz-service/internal/config"
	grpcCtrl "pvz-service/internal/controller/grpc"
	"pvz-service/internal/logger/sl"
	"pvz-service/internal/service"
	"sync"

	"google.golang.org/grpc"
)

func StartGRPCServer(cfg *config.Config, log *slog.Logger, pvzService service.PVZServiceInterface) func(*sync.WaitGroup) {

	// Создание gRPC сервера
	grpcServer := grpc.NewServer()
	pvzGrpcServer := grpcCtrl.NewPVZServer(pvzService)
	pvzGrpcServer.Register(grpcServer)

	// Запуск gRPC сервера
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.GRPC.Host, cfg.GRPC.Port))
	if err != nil {
		log.Error("failed to create gRPC server", sl.Err(err))
		return func(wg *sync.WaitGroup) {}
	}

	go func() {
		log.Info("gRPC server starting", slog.String("address", fmt.Sprintf("%s:%s", cfg.GRPC.Host, cfg.GRPC.Port)))
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("gRPC server failed", sl.Err(err))
		}
	}()

	// Graceful Stop
	return func(wg *sync.WaitGroup) {
		defer wg.Done()

		log.Info("Shutting down gRPC server")
		grpcServer.GracefulStop()
		log.Info("HTTP server gracefully stopped")
	}
}
