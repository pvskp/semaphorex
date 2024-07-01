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
	v.ConnectToServers()
	v.ClientRegisterVehicle()
	v.GetPeers()

	defer v.LeaderConn.Close()
	defer v.DiscoveryConn.Close()

	for {
		if v.IsLeader {
			v.UpdateVehicleList()
		} else {
			// is not leader
			v.ClientGetInstructions()

			if !v.CheckLeaderHealth() {
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
}
