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

	// log.Printf("GetVehiclesDirections request: %v", req)
	log.Println("returning upslice:", upSlice)
	log.Println("returning downslice:", downSlice)
	log.Println("returning leftslice:", leftSlice)
	log.Println("returning rightslice:", rightSlice)

	return &pb.GetVehiclesDirectionsResponse{
		Up:    upSlice,
		Down:  downSlice,
		Left:  leftSlice,
		Right: rightSlice,
	}, nil
}

// func (vd *VehicleDiscovery) hasLeader() bool {
// 	vd.mu.Lock()
// 	defer vd.mu.Unlock()
// 	haveLeader := false
// 	// var leader *pb.Vehicle = nil
//
// 	for _, i := range [][]*pb.Vehicle{upSlice, leftSlice, rightSlice, downSlice} {
// 		for _, v := range i {
// 			if v.IsLeader {
// 				haveLeader = true
// 				// leader = v
// 			}
// 		}
// 	}
// 	return haveLeader
//
// }

func (vd *VehicleDiscovery) HasLeader(ctx context.Context, req *pb.HasLeaderRequest) (*pb.HasLeaderResponse, error) {
	vd.mu.Lock()
	defer vd.mu.Unlock()
	haveLeader := false

	var leader *pb.Vehicle = nil

	for _, i := range [][]*pb.Vehicle{upSlice, leftSlice, rightSlice, downSlice} {
		for _, v := range i {
			if v.IsLeader {
				haveLeader = true
				leader = v
			}
		}
	}

	return &pb.HasLeaderResponse{
		HaveLeader: haveLeader,
		LeaderInfo: leader,
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

	vd.VehiclesConnected = []*pb.Vehicle{}

	log.Println("Received the following vehicles:", req.Vehicles)
	upSlice = []*pb.Vehicle{}
	downSlice = []*pb.Vehicle{}
	leftSlice = []*pb.Vehicle{}
	rightSlice = []*pb.Vehicle{}

	for _, value := range req.Vehicles {
		vd.VehiclesConnected = append(vd.VehiclesConnected, value)

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

	log.Println("upslice after, ", upSlice)
	for _, c := range vd.VehiclesConnected {
		log.Println("after received, ", c.Address, c.ShouldWalk)
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
	log.Printf("vd.VehiclesConnected: %v", vd.VehiclesConnected)
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
