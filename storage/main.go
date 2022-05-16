package main

import "dfs/storage/microservice"

func main() {
	storageMicroservice := microservice.NewStorageMicroservice()
	storageMicroservice.Setup()
	storageMicroservice.Run()
	defer storageMicroservice.Cleanup()
}
