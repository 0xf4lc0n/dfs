package main

import "dfs/auth/microservice"

func main() {
	authMicroservice := microservice.NewAuthMicroservice()
	authMicroservice.Setup()
	authMicroservice.Run()
	defer authMicroservice.Cleanup()
}
