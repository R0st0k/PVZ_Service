package main

import (
	"log"
	"os"
	"pvz-service/internal/app"
	"pvz-service/internal/config"
	"pvz-service/internal/controller/http/router"
	"pvz-service/internal/logger"
	"pvz-service/internal/logger/sl"
	"pvz-service/internal/metrics"
	"pvz-service/internal/repository"
	"pvz-service/internal/service"
	"sync"
)

func main() {
	// Load config or exit
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Init logger
	log := logger.Setup()

	// Init AuthRepo and AuthService
	authRepo, err := repository.CreateAuthRepo(cfg, log)
	if err != nil {
		log.Error("failed to init auth repo", sl.Err(err))
		os.Exit(1)
	}
	defer authRepo.CloseConnection()
	authService := service.NewAuthService(authRepo, cfg, log)

	// Init PVZRepo and PVZService
	pvzRepo, err := repository.CreatePVZRepo(cfg, log)
	if err != nil {
		log.Error("failed to init pvz repo", sl.Err(err))
		os.Exit(1)
	}
	defer pvzRepo.CloseConnection()
	pvzService := service.NewPVZService(pvzRepo, log)

	metrics := metrics.NewMetrics()

	serversStopFuncs := make([]func(*sync.WaitGroup), 0, 3)

	// Setup prometheus server
	if cfg.Prometheus.IsAble {
		log.Info("metrics server is enabled")
		serversStopFuncs = append(serversStopFuncs, app.StartMetricsServer(cfg, log, metrics))
	}

	// Setup http server router
	router := router.Setup(log, metrics, *authService, *pvzService)

	// Start http server
	serversStopFuncs = append(serversStopFuncs, app.StartHTTPServer(cfg, log, &router))

	// Setup grpc server
	if cfg.GRPC.IsAble {
		log.Info("gRPC server is enabled")
		serversStopFuncs = append(serversStopFuncs, app.StartGRPCServer(cfg, log, pvzService))
	}

	// Wait for terminate
	app.WaitForTerminate(serversStopFuncs...)
}
