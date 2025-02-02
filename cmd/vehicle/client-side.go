package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	pb "github.com/pvskp/semaphorex/pkg/coord"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	maxRetries = 5
	retryDelay = 2 * time.Second
)

var (
	leaderAddress   = ""
	SDServerAddress string
)

// Função retry que gerencia a lógica de retry
func retry(ctx context.Context, operation func(ctx context.Context) error) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = operation(ctx)
		if err == nil {
			return nil
		}

		st, ok := status.FromError(err)
		if ok && (st.Code() == codes.Unavailable || st.Code() == codes.DeadlineExceeded) {
			log.Printf("Retry %d/%d: operation failed: %v", i+1, maxRetries, err)
			time.Sleep(retryDelay)
			continue
		}

		return err
	}

	return err
}

func NewVehicle(name string) *Vehicle {

	hostname := os.Getenv("VEHICLE_NAME")
	if hostname == "" {
		log.Fatalf("VEHICLE_NAME is empty")
	}
	hostname = fmt.Sprintf("%s:%s", hostname, vehiclePort)

	return &Vehicle{
		Vehicle: &pb.Vehicle{
			Name:        name,
			Address:     hostname,
			IsLeader:    false,
			LogicalTime: 0,
			Direction:   randomDirection(),
			// Direction:  "down",
			Id:         uuid.NewString(),
			ShouldWalk: false,
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

	client := v.DiscoveryClient.(pb.VehicleDiscoveryClient)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	allPeers := []*pb.Vehicle{}

	for _, i := range [][]*pb.Vehicle{upSlice, leftSlice, rightSlice, downSlice} {
		allPeers = append(allPeers, i...)
	}

	// allPeers = append(allPeers, v.Vehicle)

	req := &pb.UpdateVehicleListRequest{
		Vehicles:    allPeers,
		LogicalTime: v.LogicalTime,
	}

	operation := func(ctx context.Context) error {
		res, err := client.UpdateVehicleList(ctx, req)
		if err != nil {
			return err
		}

		if !res.Success {
			log.Println("Error updating vehicle list")
		} else {
			// log.Println("Vehicle list updated successfully")
		}
		return nil
	}

	err := retry(ctx, operation)
	if err != nil {
		log.Printf("could not update vehicle list after %d retries: %v", maxRetries, err)
	}
}

func (v *Vehicle) ConnectToLeader() {
	v.mu.Lock()
	defer v.mu.Unlock()

	log.Printf("Connecting to leader %s", leaderAddress)
	leaderConn, err := grpc.NewClient(leaderAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("did not connect: %v", err)
	}

	v.LeaderConn = leaderConn
	v.LeaderClient = pb.NewCoordinationServiceClient(v.LeaderConn)
	log.Printf("Connected to leader %s", leaderAddress)
}

func (v *Vehicle) ConnectToVDServer() {
	v.mu.Lock()
	defer v.mu.Unlock()

	discoveryConn, err := grpc.NewClient(SDServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("did not connect: %v", err)
	}

	v.DiscoveryConn = discoveryConn
	v.DiscoveryClient = pb.NewVehicleDiscoveryClient(v.DiscoveryConn)
}

func (v *Vehicle) ClientCheckLeaderHealth() bool {
	client := v.LeaderClient.(pb.CoordinationServiceClient)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.CheckLeaderHealthRequest{
		LogicalTime: v.LogicalTime,
	}

	operation := func(ctx context.Context) error {
		res, err := client.CheckLeaderHealth(ctx, req)
		if err != nil {
			return err
		}
		log.Printf("CheckLeaderHealth response: %v", res)
		return nil
	}

	err := retry(ctx, operation)
	if err != nil {
		log.Printf("Leader's health check failed in the last %d retries: %v", maxRetries, err)
		return false
	}

	return true
}

func randomDirection() string {
	return []string{"up", "down", "left", "right"}[rand.Intn(4)]
}

func (v *Vehicle) ClientAppendPossible(dir string) bool {
	var client SpecialClient
	if _, ok := v.LeaderClient.(pb.CoordinationServiceClient); !ok {
		client = v.DiscoveryClient.(pb.VehicleDiscoveryClient)
	} else {
		client = v.LeaderClient.(pb.CoordinationServiceClient)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.AppendPossibleRequest{
		Dir: dir,
		Requester: &pb.Vehicle{
			Name:        v.Name,
			Address:     v.Address,
			IsLeader:    false,
			LogicalTime: v.LogicalTime,
			Id:          v.Id,
			Direction:   randomDirection(),
			// Direction: "up",
		}}

	for i := 0; i < maxRetries; i++ {
		res, err := client.AppendPossible(ctx, req)

		if err == nil {
			log.Printf("AppendPossible Response: %v", res)
			return res.Possible
		}

		st, ok := status.FromError(err)
		if ok && (st.Code() == codes.Unavailable || st.Code() == codes.DeadlineExceeded) {
			log.Printf("Retry %d/%d: operation failed: %v", i+1, maxRetries, err)
			time.Sleep(retryDelay)
			continue
		}

	}
	return false

}

func (v *Vehicle) HasLeader() bool {
	client := v.DiscoveryClient.(pb.VehicleDiscoveryClient)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.HasLeaderRequest{
		Requester: v.Vehicle,
	}

	for i := 0; i < maxRetries; i++ {
		res, err := client.HasLeader(ctx, req)
		if err == nil {
			log.Printf("RegisterVehicle Response: %v", res)
			leaderAddress = res.LeaderInfo.Address
			return res.HaveLeader
		}

		st, ok := status.FromError(err)
		if ok && (st.Code() == codes.Unavailable || st.Code() == codes.DeadlineExceeded) {
			log.Printf("Retry %d/%d: operation failed: %v", i+1, maxRetries, err)
			time.Sleep(retryDelay)
			continue
		}
	}
	return false
}

func (v *Vehicle) ClientRegisterVehicle() {
	client := v.LeaderClient

	if client == nil {
		client = v.DiscoveryClient
		log.Println("Registering on discovery server")
	} else {
		log.Println("Registering on leader server")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rDir := randomDirection()
	// rDir := "up"

	fmt.Printf("Trying to append on %s\n", rDir)
	for !(v.ClientAppendPossible(rDir)) {
		log.Println("Failed to append on position, retrying...")
		rDir = randomDirection()
		fmt.Printf("Trying to append on %s\n", rDir)
	}

	req := &pb.RegisterVehicleRequest{Vehicle: &pb.Vehicle{
		Name:        v.Name,
		Address:     v.Address,
		IsLeader:    false,
		LogicalTime: v.LogicalTime,
		Id:          v.Id,
		Direction:   rDir,
	}}

	operation := func(ctx context.Context) error {
		res, err := client.RegisterVehicle(ctx, req)
		if err != nil {
			return err
		}
		log.Printf("RegisterVehicle Response: %v", res)
		return nil
	}

	err := retry(ctx, operation)
	if err != nil {
		log.Printf("could not register vehicle after %d retries: %v", maxRetries, err)
	}

	// downSlice = append(downSlice, v.Vehicle)

	switch rDir {
	case "up":
		upSlice = append(upSlice, v.Vehicle)
	case "down":
		downSlice = append(downSlice, v.Vehicle)
	case "left":
		leftSlice = append(leftSlice, v.Vehicle)
	case "right":
		rightSlice = append(rightSlice, v.Vehicle)
	}

}

func (v *Vehicle) ClientGetInstructions() {
	client := v.LeaderClient.(pb.CoordinationServiceClient)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.GetInstructionsRequest{
		VehicleAddress: v.Address,
		LogicalTime:    v.LogicalTime,
	}

	operation := func(ctx context.Context) error {
		res, err := client.GetInstructions(ctx, req)
		if err != nil {
			return err
		}
		log.Printf("Got instruction: %v", res.Instruction)
		v.LogicalTime = res.LogicalTime
		return nil
	}

	err := retry(ctx, operation)
	if err != nil {
		log.Printf("could not get instructions after %d retries: %v", maxRetries, err)
	}
}

func (v *Vehicle) GetPeers() {
	v.mu.Lock()
	defer v.mu.Unlock()

	client := v.DiscoveryClient.(pb.VehicleDiscoveryClient)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	operation := func(ctx context.Context) error {
		res, err := client.ListRegisteredVehicles(ctx, &pb.ListRegisteredVehiclesRequest{
			Requester: v.Vehicle,
		})

		if err != nil {
			return err
		}

		v.peers = res.Vehicles

		for _, p := range v.peers {
			if p.IsLeader && leaderAddress == "" {
				leaderAddress = p.Address
				log.Println("Found leader: ", leaderAddress)
			}
		}

		// log.Printf("ListRegisteredVehicles Response: %v", res)
		// for _, value := range res.Vehicles {
		// 	if value.IsLeader && leaderAddress == "" {
		// 		leaderAddress = value.Address
		// 		log.Println("Found leader: ", leaderAddress)
		// 	}
		//
		// 	if value.Address != v.Address {
		// 		v.peers = append(v.peers, value)
		// 		log.Println("Found a peer!")
		// 	}
		// }

		return nil
	}

	err := retry(ctx, operation)

	if err != nil {
		log.Printf("could not get peers after %d retries: %v", maxRetries, err)
	}
}

func (v *Vehicle) InitiateElection() {
	log.Println("Initiating leader election")
	v.mu.Lock()
	v.LogicalTime++
	electionResultCh := make(chan *pb.ElectLeaderResponse, len(v.peers))
	var wg sync.WaitGroup
	v.mu.Unlock()

	log.Println("My peers are: ", v.peers)
	for _, peer := range v.peers {
		wg.Add(1)
		log.Printf("Trying to connect to peer %v...", peer)
		go func(peer *pb.Vehicle) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			req := &pb.ElectLeaderRequest{
				RequesterId: v.Id,
				LogicalTime: v.LogicalTime,
			}

			peerConn, err := grpc.NewClient(peer.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Printf("did not connect to peer %s: %v", peer.Address, err)
				return
			}
			defer peerConn.Close()

			peerClient := pb.NewCoordinationServiceClient(peerConn)

			operation := func(ctx context.Context) error {
				res, err := peerClient.ElectLeader(ctx, req)
				if err != nil {
					return err
				}
				electionResultCh <- res
				return nil
			}

			err = retry(ctx, operation)
			if err != nil {
				log.Printf("Error during leader election after retries: %v", err)
			}
		}(peer)
	}

	wg.Wait()
	close(electionResultCh)

	v.mu.Lock()
	defer v.mu.Unlock()

	var highestLogicalTime int64
	var newLeaderId string

	for res := range electionResultCh {
		log.Printf("Received response: %v", res)
		if res.LeaderTime > highestLogicalTime {
			highestLogicalTime = res.LeaderTime
			newLeaderId = res.LeaderId
			leaderAddress = res.LeaderAddress
		}
	}

	if newLeaderId == v.Id {
		v.IsLeader = true
		log.Printf("Vehicle %s is the new coordinator", v.Address)
	} else {
		v.IsLeader = false
		log.Printf("Vehicle %s is not the leader. New leader is %s, from address %s", v.Id, newLeaderId, leaderAddress)
	}
}
