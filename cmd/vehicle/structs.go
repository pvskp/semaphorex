package main

import (
	"sync"

	pb "github.com/pvskp/semaphorex/pkg/coord"
	"google.golang.org/grpc"
	// "google.golang.org/grpc/credentials/insecure"
)

type Vehicle struct {
	*pb.Vehicle
	pb.UnimplementedCoordinationServiceServer
	LeaderConn      *grpc.ClientConn
	DiscoveryConn   *grpc.ClientConn
	LeaderClient    pb.CoordinationServiceClient
	DiscoveryClient pb.VehicleDiscoveryClient
	peers           []string
	mu              sync.Mutex
}
