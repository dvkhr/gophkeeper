package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/dvkhr/gophkeeper/pb"
	"github.com/dvkhr/gophkeeper/pkg/logger"
	"github.com/dvkhr/gophkeeper/server/internal/api"
	"github.com/dvkhr/gophkeeper/server/internal/auth"
	"github.com/dvkhr/gophkeeper/server/internal/config"
	"github.com/dvkhr/gophkeeper/server/internal/db"
	"github.com/dvkhr/gophkeeper/server/internal/repository"
	"github.com/dvkhr/gophkeeper/server/internal/service"
	"google.golang.org/grpc"
)

func main() {
	// Инициализируем логгер
	if err := logger.InitLogger("/home/max/go/src/GophKeeper/configs/logger.yaml"); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Парсим флаги командной строки
	config.ParseFlags()
	logger.Logg.Info("Using config file", "path", config.ConfigFile)

	// Загрузка основной конфигурации
	cfg, err := config.Load(config.ConfigFile)
	if err != nil {
		logger.Logg.Error("Failed to load config", "error", err)
		panic(err)
	}

	// Подключение к базе данных
	dbConn, err := db.Connect(cfg.Database.DSN)
	if err != nil {
		logger.Logg.Error("Failed to connect to DB", "error", err)
		return
	}
	defer dbConn.Close()

	if err := db.ApplyMigrations(dbConn); err != nil {
		logger.Logg.Error("Failed to apply migrations", "error", err)
		return
	}

	logger.Logg.Info("Database is ready. Starting server...")

	repo := repository.NewPostgresRepository(dbConn)
	service := service.New(repo, cfg)
	server := api.NewKeeperServer(service)

	// Подготовка gRPC сервера
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		logger.Logg.Error("Failed to listen", "error", err)
		panic(err)
	}
	interceptor := auth.AuthInterceptor(*cfg, repo)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor),
	)

	pb.RegisterKeeperServiceServer(grpcServer, server)
	logger.Logg.Info("Starting gRPC server",
		"port", cfg.Server.Port,
		"mode", cfg.Server.Mode,
		"dsn", cfg.Database.DSN,
		"jwt_ttl_Hours", cfg.Auth.JWTTTLHours,
		"jwt_ttl_Minutes", cfg.Auth.JWTTTLMinutes,
	)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Logg.Error("Failed to serve", "error", err)
			panic(err)
		}
	}()

	logger.Logg.Info("Server is running...")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	logger.Logg.Info("Shutting down server...")
	grpcServer.GracefulStop()
}
