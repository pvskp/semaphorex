package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/pvskp/semaphorex/pkg/coord"
	"google.golang.org/grpc"
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

func (v *Vehicle) UpdateVehicleStatus(ctx context.Context, req *pb.UpdateVehicleStatusRequest) (*pb.UpdateVehicleStatusResponse, error) {
	log.Printf("UpdateVehicleStatus request: %v", req)
	return &pb.UpdateVehicleStatusResponse{
		Success: true,
		Message: "Vehicle status updated successfully",
	}, nil
}

func (v *Vehicle) GetInstructions(ctx context.Context, req *pb.GetInstructionsRequest) (*pb.GetInstructionsResponse, error) {
	log.Printf("GetInstructions request: %v", req)
	return &pb.GetInstructionsResponse{
		Instruction: "proceed",
		LogicalTime: req.LogicalTime,
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
			LeaderId:   v.Id,
			LeaderTime: v.LogicalTime,
		}, nil
	}

	return &pb.ElectLeaderResponse{
		LeaderId:   req.RequesterId,
		LeaderTime: req.LogicalTime,
	}, nil
}
