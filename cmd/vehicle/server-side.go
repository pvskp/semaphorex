package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/pvskp/semaphorex/pkg/coord"
	"google.golang.org/grpc"
)

var (
	upSlice    = []*pb.Vehicle{}
	downSlice  = []*pb.Vehicle{}
	leftSlice  = []*pb.Vehicle{}
	rightSlice = []*pb.Vehicle{}
)

func (v *Vehicle) StartServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", vehiclePort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCoordinationServiceServer(s, v)
	log.Printf("Server with UUID %s was registred on address %s\n", v.Id, v.Address)
	log.Printf("Server listening on port %v", vehiclePort)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
}

func (v *Vehicle) RegisterVehicle(ctx context.Context, req *pb.RegisterVehicleRequest) (*pb.RegisterVehicleResponse, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	log.Printf("RegisterVehicle request: %v", req)
	v.peers = append(v.peers, req.Vehicle)

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

func (v *Vehicle) UpdateVehicleStatus(ctx context.Context, req *pb.UpdateVehicleStatusRequest) (*pb.UpdateVehicleStatusResponse, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if req.Vehicle.LogicalTime > v.LogicalTime {
		v.LogicalTime = req.Vehicle.LogicalTime
	}

	v.LogicalTime++

	log.Printf("UpdateVehicleStatus request: %v", req)
	return &pb.UpdateVehicleStatusResponse{
		Success:     true,
		Message:     "Vehicle status updated successfully",
		LogicalTime: v.LogicalTime,
	}, nil
}

func (v *Vehicle) CheckLeaderHealth(ctx context.Context, req *pb.CheckLeaderHealthRequest) (*pb.CheckLeaderHealthResponse, error) {
	if req.LogicalTime > v.LogicalTime {
		v.LogicalTime = req.LogicalTime
	}

	v.LogicalTime++

	return &pb.CheckLeaderHealthResponse{
		Success:     true,
		LogicalTime: v.LogicalTime,
	}, nil
}

func (v *Vehicle) GetInstructions(ctx context.Context, req *pb.GetInstructionsRequest) (*pb.GetInstructionsResponse, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if req.LogicalTime > v.LogicalTime {
		v.LogicalTime = req.LogicalTime
	}

	v.LogicalTime++

	log.Printf("GetInstructions request: %v", req)
	return &pb.GetInstructionsResponse{
		Instruction: "proceed",
		LogicalTime: v.LogicalTime,
	}, nil
}

func (v *Vehicle) ElectLeader(ctx context.Context, req *pb.ElectLeaderRequest) (*pb.ElectLeaderResponse, error) {
	log.Printf("ElectLeader request: %v", req)
	v.mu.Lock()
	defer v.mu.Unlock()

	if req.LogicalTime > v.LogicalTime {
		v.LogicalTime = req.LogicalTime
	}

	v.LogicalTime++

	// Lógica para eleger um líder
	if v.Id > req.RequesterId {
		return &pb.ElectLeaderResponse{
			LeaderId:      v.Id,
			LeaderTime:    v.LogicalTime,
			LeaderAddress: v.Address,
		}, nil
	}

	return &pb.ElectLeaderResponse{
		LeaderId:      req.RequesterId,
		LeaderTime:    req.LogicalTime,
		LeaderAddress: req.RequesterAddress,
	}, nil
}
