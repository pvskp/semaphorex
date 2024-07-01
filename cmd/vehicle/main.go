package main

import (
	"time"
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
