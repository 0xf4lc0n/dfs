package main

import "dfs/storage/microservice"

func main() {
	storageMicroservice := microservice.NewStorageMicroservice()
	storageMicroservice.Setup()
	defer storageMicroservice.Cleanup()
	storageMicroservice.Run()
}
