package main

import (
	"context"
	"log"
	"net"

	pb "github.com/pvskp/semaphorex/pkg/coord"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

type server struct {
	pb.UnimplementedCoordinationServiceServer
	// Adicione campos adicionais para gerenciar o estado do servidor, se necessário
}

func (s *server) RegisterVehicle(ctx context.Context, req *pb.RegisterVehicleRequest) (*pb.RegisterVehicleResponse, error) {
	log.Printf("RegisterVehicle request: %v", req)
	// Adicione lógica para registrar o veículo
	return &pb.RegisterVehicleResponse{
		Success: true,
		Message: "Vehicle registered successfully",
	}, nil
}

func (s *server) UpdateVehicleStatus(ctx context.Context, req *pb.UpdateVehicleStatusRequest) (*pb.UpdateVehicleStatusResponse, error) {
	log.Printf("UpdateVehicleStatus request: %v", req)
	// Adicione lógica para atualizar o status do veículo
	return &pb.UpdateVehicleStatusResponse{
		Success: true,
		Message: "Vehicle status updated successfully",
	}, nil
}

func (s *server) GetInstructions(ctx context.Context, req *pb.GetInstructionsRequest) (*pb.GetInstructionsResponse, error) {
	log.Printf("GetInstructions request: %v", req)
	// Adicione lógica para fornecer instruções ao veículo
	return &pb.GetInstructionsResponse{
		Instruction: "proceed",
		LogicalTime: req.LogicalTime,
	}, nil
}

func (s *server) SyncLogicalClock(ctx context.Context, req *pb.SyncLogicalClockRequest) (*pb.SyncLogicalClockResponse, error) {
	log.Printf("SyncLogicalClock request: %v", req)
	// Adicione lógica para sincronizar o relógio lógico
	return &pb.SyncLogicalClockResponse{
		UpdatedTime: req.LocalTime + 1,
	}, nil
}

func (s *server) ElectLeader(ctx context.Context, req *pb.ElectLeaderRequest) (*pb.ElectLeaderResponse, error) {
	log.Printf("ElectLeader request: %v", req)
	// Adicione lógica para eleger um líder
	return &pb.ElectLeaderResponse{
		LeaderId:   req.VehicleId,
		LeaderTime: req.LogicalTime,
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCoordinationServiceServer(s, &server{})
	log.Printf("Server listening on port %v", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
