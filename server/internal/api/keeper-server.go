package api

import (
	"context"

	"github.com/dvkhr/gophkeeper/pb"
	"github.com/dvkhr/gophkeeper/pkg/logger"
)

type KeeperServer struct {
	pb.UnimplementedKeeperServiceServer
}

func (s *KeeperServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {
	logger.Logg.Info("Register request: %v", req.Login)
	return &pb.AuthResponse{
		AccessToken:  "fake-jwt-token",
		RefreshToken: "fake-refresh-token",
	}, nil
}

func (s *KeeperServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	logger.Logg.Info("Login request: %v", req.Login)
	return &pb.AuthResponse{
		AccessToken:  "fake-jwt-token",
		RefreshToken: "fake-refresh-token",
	}, nil
}

func (s *KeeperServer) StoreData(ctx context.Context, req *pb.StoreDataRequest) (*pb.StatusResponse, error) {
	logger.Logg.Info("Storing data: %v", req.Record.Type)
	return &pb.StatusResponse{Success: true, Message: "Stored"}, nil
}

func (s *KeeperServer) GetData(ctx context.Context, req *pb.GetDataRequest) (*pb.DataResponse, error) {
	logger.Logg.Info("Get data by type: %v", req.Type)
	return &pb.DataResponse{}, nil
}

func (s *KeeperServer) SyncData(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {
	logger.Logg.Info("Syncing %d records", len(req.Records))
	return &pb.SyncResponse{}, nil
}
