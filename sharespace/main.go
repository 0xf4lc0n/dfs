package main

import "dfs/sharespace/microservice"

func main() {
	shareMicroservice := microservice.NewShareSpaceMicroservice()
	shareMicroservice.Setup()
	shareMicroservice.Run()
	defer shareMicroservice.Cleanup()
}
