package main

import (
	"log"
	"os"
	"time"
)

var (
	vehiclePort string
)

func main() {
	v := NewVehicle("porshce")
	v.StartServer()
	v.ConnectToVDServer()

	log.Println("Waiting for peers...")

	v.GetPeers()
	if len(v.peers) == 0 {
		v.IsLeader = true
		log.Printf("%s is the leader", v.Address)
	}

	if leaderAddress != "" {
		v.ConnectToLeader()
	}

	v.ClientRegisterVehicle()

	defer v.LeaderConn.Close()
	defer v.DiscoveryConn.Close()

	for {
		if v.IsLeader {
			v.UpdateVehicleList()
		} else {
			// is not leader
			if v.LeaderConn == nil {
				v.ConnectToLeader()
			}
			log.Printf("%s is not the leader. Getting instructions from %s...", v.Address, leaderAddress)
			v.ClientGetInstructions()

			if !v.ClientCheckLeaderHealth() {
				v.InitiateElection()
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func init() {
	vehiclePort = os.Getenv("VEHICLE_PORT")
	if vehiclePort == "" {
		log.Fatalf("VEHICLE_PORT is empty")
	}
	log.Printf("VehiclePort: %s", vehiclePort)

	SDServerAddress = os.Getenv("SD_SERVER_ADDRESS")
	if SDServerAddress == "" {
		log.Fatalf("SD_SERVER_ADDRESS is empty")
	}
	log.Printf("SDServerAddress: %s", SDServerAddress)
}
