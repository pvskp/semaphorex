package main

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	pb "github.com/pvskp/semaphorex/pkg/coord"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	SDServerAddress = "localhost:50053"
	vehiclePort     = "50052"
)

var (
	leaderAddress = ""
)

func NewVehicle(name string) *Vehicle {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal("Couldn't get hostname")
	}

	return &Vehicle{
		Vehicle: &pb.Vehicle{
			Name:        name,
			Id:          uuid.NewString(),
			Address:     hostname,
			IsLeader:    false,
			LogicalTime: 0,
			PosX:        0,
			PosY:        0,
		},

		peers:           []string{},
		LeaderConn:      nil,
		DiscoveryConn:   nil,
		LeaderClient:    nil,
		DiscoveryClient: nil,
		mu:              sync.Mutex{},
	}
}

func (v *Vehicle) ConnectToServers() {
	leaderConn, err := grpc.NewClient(leaderAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	discoveryConn, err := grpc.NewClient(SDServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	v.LeaderConn = leaderConn
	v.DiscoveryConn = discoveryConn

	v.LeaderClient = pb.NewCoordinationServiceClient(v.LeaderConn)
	v.DiscoveryClient = pb.NewVehicleDiscoveryClient(v.DiscoveryConn)
}

// CheckLeaderHealth checks if the leader is healthy (reachable, for example).
// If it's not, returns false.
func (v *Vehicle) CheckLeaderHealth() bool {
	return false
}

func (v *Vehicle) ClientRegisterVehicle() {
	client := v.DiscoveryClient
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &pb.RegisterVehicleRequest{Vehicle: &pb.Vehicle{
		Name:        v.Name,
		Address:     v.Address,
		PosX:        v.PosX,
		PosY:        v.PosY,
		IsLeader:    false,
		LogicalTime: v.LogicalTime,
	}}

	res, err := client.RegisterVehicle(ctx, req)
	if err != nil {
		log.Fatalf("could not register vehicle: %v", err)
	}
	log.Printf("RegisterVehicle Response: %v", res)
}

func (v *Vehicle) ClientGetInstructions() {
	client := v.LeaderClient
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &pb.GetInstructionsRequest{VehicleAddress: v.Address}
	res, err := client.GetInstructions(ctx, req)
	if err != nil {
		log.Fatalf("could not get instructions: %v", err)
	}
	log.Printf("GetInstructions Response: %v", res)
	v.LogicalTime = res.LogicalTime
}

func (v *Vehicle) GetPeers() {
	client := v.DiscoveryClient

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.ListRegisteredVehicles(ctx, &pb.ListRegisteredVehiclesRequest{})
	if err != nil {
		log.Fatalf("could not get instructions: %v", err)
	}

	log.Printf("ListRegisteredVehicles Response: %v", res)
	for _, value := range res.Vehicles {
		if value.IsLeader {
			leaderAddress = value.Address
		}

		if value.Address != v.Address {
			v.peers = append(v.peers, value.Address)
		}
	}

}

func (v *Vehicle) InitiateElection() {
	log.Println("Initiating leader election")
	for _, peer := range v.peers {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		req := &pb.ElectLeaderRequest{
			RequesterId: v.Id,
			LogicalTime: v.LogicalTime,
		}

		peerConn, err := grpc.NewClient(peer, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer peerConn.Close()
		if err != nil {
			log.Fatalf("did not connect to peer %s: %v", peer, err)
		}

		peerClient := pb.NewCoordinationServiceClient(v.LeaderConn)

		res, err := peerClient.ElectLeader(ctx, req)
		if err != nil {
			log.Printf("Error during leader election: %v", err)
			continue
		}
		log.Printf("Leader election response: %v", res)
		if res.LeaderId != v.Id {
			v.IsLeader = false
			return
		}
	}
	v.IsLeader = true
	log.Printf("Vehicle %s is the new coordinator", v.Id)
}
