package main

import "dfs/share/microservice"

func main() {
	shareMicroservice := microservice.NewShareMicroservice()
	shareMicroservice.Setup()
	shareMicroservice.Run()
	defer shareMicroservice.Cleanup()
}
