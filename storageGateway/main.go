package main

import "dfs/storageGateway/microservice"

func main() {
	gatewayMicroservice := microservice.NewGatewayMicroservice()
	gatewayMicroservice.Setup()
	defer gatewayMicroservice.Cleanup()
	gatewayMicroservice.Run()
}
