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

func (v *Vehicle) AppendPossible(ctx context.Context, req *pb.AppendPossibleRequest) (*pb.AppendPossibleResponse, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

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

func (v *Vehicle) RegisterVehicle(ctx context.Context, req *pb.RegisterVehicleRequest) (*pb.RegisterVehicleResponse, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	log.Printf("RegisterVehicle request: %v", req)

	// Atualizar as direções dos veículos
	switch req.Vehicle.Direction {
	case "up":
		for idx := range upSlice {
			upSlice[idx].ShouldWalk = true
			log.Println("upSlice should walk: ", upSlice[idx].ShouldWalk)
		}
		upSlice = append(upSlice, req.Vehicle)
	case "down":
		for idx := range downSlice {
			downSlice[idx].ShouldWalk = true
			log.Println("downSlice should walk: ", downSlice[idx].ShouldWalk)
		}
		downSlice = append(downSlice, req.Vehicle)
	case "left":
		for idx := range leftSlice {
			leftSlice[idx].ShouldWalk = true
			log.Println("leftSlice should walk: ", leftSlice[idx].ShouldWalk)
		}
		leftSlice = append(leftSlice, req.Vehicle)
	case "right":
		for idx := range rightSlice {
			rightSlice[idx].ShouldWalk = true
			log.Println("rightSlice should walk: ", rightSlice[idx].ShouldWalk)
		}
		rightSlice = append(rightSlice, req.Vehicle)
	}

	v.peers = append(v.peers, req.Vehicle)
	log.Println("peers list:", v.peers)

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

	// log.Printf("GetInstructions request: %v", req)
	return &pb.GetInstructionsResponse{
		Instruction: randomDirection(),
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
