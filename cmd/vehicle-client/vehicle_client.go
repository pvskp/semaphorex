package main

import (
	"context"
	"log"
	"time"

	pb "github.com/pvskp/semaphorex/pkg/coord"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	serverAddress = "localhost:50051"
)

type vehicle struct {
	name        string
	id          int
	vType       string
	logicalTime int64
}

func main() {
	conn, err := grpc.NewClient(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewCoordinationServiceClient(conn)
	v := NewVehicle("porshce", 1, "car")
	v.registerVehicle(client, v.name, v.vType)
	v.updateVehicleStatus(client, v.name, 37.7749, -122.4194, "turn_left", 1)
	v.getInstructions(client, v.name)
	v.syncLogicalClock(client, v.name, 1)
	v.electLeader(client, v.name, 1)
}

func NewVehicle(name string, id int, vType string) *vehicle {
	return &vehicle{name, id, vType, 0}
}

func (v *vehicle) registerVehicle(client pb.CoordinationServiceClient, vehicleID, vehicleType string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &pb.RegisterVehicleRequest{VehicleId: vehicleID, VehicleType: vehicleType}
	res, err := client.RegisterVehicle(ctx, req)
	if err != nil {
		log.Fatalf("could not register vehicle: %v", err)
	}
	log.Printf("RegisterVehicle Response: %v", res)
}

func (v *vehicle) updateVehicleStatus(client pb.CoordinationServiceClient, vehicleID string, latitude, longitude float64, intention string, logicalTime int64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &pb.UpdateVehicleStatusRequest{
		VehicleId:   vehicleID,
		Latitude:    latitude,
		Longitude:   longitude,
		Intention:   intention,
		LogicalTime: logicalTime,
	}
	res, err := client.UpdateVehicleStatus(ctx, req)
	if err != nil {
		log.Fatalf("could not update vehicle status: %v", err)
	}
	log.Printf("UpdateVehicleStatus Response: %v", res)

	v.logicalTime = res.LogicalTime
}

func (v *vehicle) getInstructions(client pb.CoordinationServiceClient, vehicleID string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &pb.GetInstructionsRequest{VehicleId: vehicleID}
	res, err := client.GetInstructions(ctx, req)
	if err != nil {
		log.Fatalf("could not get instructions: %v", err)
	}
	log.Printf("GetInstructions Response: %v", res)
	v.logicalTime = res.LogicalTime
}

func (v *vehicle) syncLogicalClock(client pb.CoordinationServiceClient, vehicleID string, localTime int64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &pb.SyncLogicalClockRequest{VehicleId: vehicleID, LocalTime: localTime}
	res, err := client.SyncLogicalClock(ctx, req)
	if err != nil {
		log.Fatalf("could not sync logical clock: %v", err)
	}
	log.Printf("SyncLogicalClock Response: %v", res)
	v.logicalTime = res.UpdatedTime
}

func (v *vehicle) electLeader(client pb.CoordinationServiceClient, vehicleID string, logicalTime int64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &pb.ElectLeaderRequest{VehicleId: vehicleID, LogicalTime: logicalTime}
	res, err := client.ElectLeader(ctx, req)
	if err != nil {
		log.Fatalf("could not elect leader: %v", err)
	}
	log.Printf("ElectLeader Response: %v", res)
}
