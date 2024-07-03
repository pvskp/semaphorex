package main

import (
	"context"
	"log"
	"net"
	"sync"

	pb "github.com/pvskp/semaphorex/pkg/coord"
	"google.golang.org/grpc"
)

var (
	upSlice    = []*pb.Vehicle{}
	downSlice  = []*pb.Vehicle{}
	leftSlice  = []*pb.Vehicle{}
	rightSlice = []*pb.Vehicle{}
)

type VehicleDiscovery struct {
	pb.UnimplementedVehicleDiscoveryServer
	Port              int
	VehiclesConnected []*pb.Vehicle
	mu                sync.Mutex
}

func NewVehicleDiscovery(port int) *VehicleDiscovery {
	return &VehicleDiscovery{
		Port:              port,
		VehiclesConnected: []*pb.Vehicle{},
		mu:                sync.Mutex{},
	}

}

func (vd *VehicleDiscovery) GetVehiclesDirections(ctx context.Context, req *pb.GetVehiclesDirectionsRequest) (*pb.GetVehiclesDirectionsResponse, error) {
	vd.mu.Lock()
	defer vd.mu.Unlock()

	log.Printf("GetVehiclesDirections request: %v", req)
	return &pb.GetVehiclesDirectionsResponse{
		Up:    upSlice,
		Down:  downSlice,
		Left:  leftSlice,
		Right: rightSlice,
	}, nil
}

func (vd *VehicleDiscovery) AppendPossible(ctx context.Context, req *pb.AppendPossibleRequest) (*pb.AppendPossibleResponse, error) {
	vd.mu.Lock()
	defer vd.mu.Unlock()

	var poss bool = false

	switch req.Dir {
	case "up":
		poss = len(upSlice) < 2
	case "down":
		poss = len(downSlice) < 2
	case "left":
		poss = len(leftSlice) < 2
	case "right":
		poss = len(rightSlice) < 2
	}

	log.Printf("AppendPossible : %v", poss)

	return &pb.AppendPossibleResponse{
		Possible: poss,
	}, nil
}

func (vd *VehicleDiscovery) UpdateVehicleList(ctx context.Context, req *pb.UpdateVehicleListRequest) (*pb.UpdateVehicleListResponse, error) {
	vd.mu.Lock()
	defer vd.mu.Unlock()

	for _, value := range req.Vehicles {
		switch value.Direction {
		case "up":
			upSlice = append(upSlice, value)
		case "down":
			downSlice = append(downSlice, value)
		case "left":
			leftSlice = append(leftSlice, value)
		case "right":
			rightSlice = append(rightSlice, value)
		}
	}
	return &pb.UpdateVehicleListResponse{
		Success: true,
	}, nil
}

func (vd *VehicleDiscovery) RegisterVehicle(ctx context.Context, req *pb.RegisterVehicleRequest) (*pb.RegisterVehicleResponse, error) {
	vd.mu.Lock()
	defer vd.mu.Unlock()

	log.Printf("RegisterVehicle request: %v", req)
	vd.VehiclesConnected = append(vd.VehiclesConnected, req.Vehicle)

	// Atualizar as direções dos veículos
	switch req.Vehicle.Direction {
	case "up":
		upSlice = append(upSlice, req.Vehicle)
	case "down":
		downSlice = append(downSlice, req.Vehicle)
	case "left":
		leftSlice = append(leftSlice, req.Vehicle)
	case "right":
		rightSlice = append(rightSlice, req.Vehicle)
	}

	log.Printf("car registered on %s array", req.Vehicle.Direction)

	return &pb.RegisterVehicleResponse{
		Success: true,
		Message: "Vehicle registered successfully",
	}, nil
}

func (vd *VehicleDiscovery) ListRegisteredVehicles(ctx context.Context, req *pb.ListRegisteredVehiclesRequest) (*pb.ListRegisteredVehiclesResponse, error) {
	vd.mu.Lock()
	defer vd.mu.Unlock()

	log.Printf("ListRegisteredVehicles request from address %s, id %s", req.Requester.Address, req.Requester.Id)
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
