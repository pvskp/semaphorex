package main

import (
	"context"
	"log"
	"net"

	pb "github.com/pvskp/semaphorex/pkg/coord"
	"google.golang.org/grpc"
)

type VehicleDiscovery struct {
	pb.UnimplementedVehicleDiscoveryServer
	Port              int
	VehiclesConnected []*pb.Vehicle
}

func NewVehicleDiscovery(port int) *VehicleDiscovery {
	return &VehicleDiscovery{
		Port:              port,
		VehiclesConnected: []*pb.Vehicle{},
	}

}

func (vd VehicleDiscovery) RegisterVehicle(ctx context.Context, req *pb.RegisterVehicleRequest) (*pb.RegisterVehicleResponse, error) {
	log.Printf("RegisterVehicle request: %v", req)
	return &pb.RegisterVehicleResponse{
		Success: true,
		Message: "",
	}, nil
}

func (vd VehicleDiscovery) ListRegisteredVehicles(ctx context.Context, req *pb.ListRegisteredVehiclesRequest) (*pb.ListRegisteredVehiclesResponse, error) {
	log.Printf("ListRegisteredVehicles request: %v", req)
	return &pb.ListRegisteredVehiclesResponse{
		Vehicles: vd.VehiclesConnected,
	}, nil
}

func main() {
	sd := NewVehicleDiscovery(8001)
	lis, err := net.Listen("tcp", ":8001")
	if err != nil {
		log.Fatalf("Couldn't listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterVehicleDiscoveryServer(s, sd)

	log.Printf("Server listening on port %d", sd.Port)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
