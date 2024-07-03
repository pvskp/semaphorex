package main

import (
	"context"
	"sync"

	pb "github.com/pvskp/semaphorex/pkg/coord"
	"google.golang.org/grpc"
	// "google.golang.org/grpc/credentials/insecure"
)

type SpecialClient interface {
	RegisterVehicle(ctx context.Context, in *pb.RegisterVehicleRequest, opts ...grpc.CallOption) (*pb.RegisterVehicleResponse, error)
	AppendPossible(ctx context.Context, in *pb.AppendPossibleRequest, opts ...grpc.CallOption) (*pb.AppendPossibleResponse, error)
}

type Vehicle struct {
	*pb.Vehicle
	pb.UnimplementedCoordinationServiceServer
	LeaderConn    *grpc.ClientConn
	DiscoveryConn *grpc.ClientConn
	// LeaderClient    pb.CoordinationServiceClient
	// DiscoveryClient pb.VehicleDiscoveryClient
	LeaderClient    SpecialClient
	DiscoveryClient SpecialClient
	peers           []*pb.Vehicle
	mu              sync.Mutex
}
