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
)

var (
	leaderAddress = ""
)

func NewVehicle(name string) *Vehicle {
	// hostname, err := os.Hostname()
	hostname := os.Getenv("HOSTNAME")
	if hostname == "" {
		log.Fatalf("Hostname is empty")
	}

	// if err != nil {
	// 	log.Fatal("Couldn't get hostname")
	// }

	return &Vehicle{
		Vehicle: &pb.Vehicle{
			Name:        name,
			Id:          uuid.NewString(),
			Address:     hostname,
			IsLeader:    false,
			LogicalTime: 0,
		},

		peers:           []*pb.Vehicle{},
		LeaderConn:      nil,
		DiscoveryConn:   nil,
		LeaderClient:    nil,
		DiscoveryClient: nil,
		mu:              sync.Mutex{},
	}
}

func (v *Vehicle) UpdateVehicleList() {
	v.mu.Lock()
	defer v.mu.Unlock()

	client := v.DiscoveryClient
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &pb.UpdateVehicleListRequest{
		Vehicles: v.peers,
	}

	res, err := client.UpdateVehicleList(ctx, req)
	if err != nil {
		log.Fatalf("could not update vehicle list: %v", err)
	}

	if !res.Success {
		log.Println("Error updating vehicle list")
	}

	log.Println("Vehicle list updated successfully")
}

func (v *Vehicle) ConnectToServers() {
	v.mu.Lock()
	defer v.mu.Unlock()

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
	v.mu.Lock()
	defer v.mu.Unlock()

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
			v.peers = append(v.peers, value)
		}
	}
}

func (v *Vehicle) InitiateElection() {
	log.Println("Initiating leader election")
	v.mu.Lock()
	v.LogicalTime++
	electionResultCh := make(chan *pb.ElectLeaderResponse, len(v.peers))
	var wg sync.WaitGroup
	v.mu.Unlock()

	for _, peer := range v.peers {
		wg.Add(1)
		go func(peer *pb.Vehicle) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			req := &pb.ElectLeaderRequest{
				RequesterId: v.Id,
				LogicalTime: v.LogicalTime,
			}

			peerConn, err := grpc.Dial(peer.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Printf("did not connect to peer %s: %v", peer.Address, err)
				return
			}
			defer peerConn.Close()

			peerClient := pb.NewCoordinationServiceClient(peerConn)
			res, err := peerClient.ElectLeader(ctx, req)
			if err != nil {
				log.Printf("Error during leader election: %v", err)
				return
			}
			electionResultCh <- res
		}(peer)
	}

	wg.Wait()
	close(electionResultCh)

	v.mu.Lock()
	defer v.mu.Unlock()

	var highestLogicalTime int64
	var newLeaderId string

	for res := range electionResultCh {
		if res.LeaderTime > highestLogicalTime {
			highestLogicalTime = res.LeaderTime
			newLeaderId = res.LeaderId
		}
	}

	if newLeaderId == v.Id {
		v.IsLeader = true
		log.Printf("Vehicle %s is the new coordinator", v.Id)
	} else {
		v.IsLeader = false
		leaderAddress = newLeaderId
		log.Printf("Vehicle %s is not the leader. New leader is %s", v.Id, newLeaderId)
	}
}
