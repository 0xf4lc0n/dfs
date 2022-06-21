package config

import (
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

type Config struct {
	IpAddress          string
	Port               uint64
	GRpcPort           uint64
	FullAddress        string
	DbConnectionString string
	FileStoragePath    string
}

func Create() *Config {
	cfg := &Config{}
	var cliArgs CliArgs
	kong.Parse(&cliArgs)

	err := godotenv.Load(".env")

	if err != nil {
		log.Printf("Cannot load .env file")
	}

	if cliArgs.IpAddress != "" {
		cfg.IpAddress = cliArgs.IpAddress
	} else {
		cfg.IpAddress = os.Getenv("IP_ADDRESS")
	}

	if cliArgs.Port != 0 {
		cfg.Port = cliArgs.Port
	} else {
		port, err := strconv.ParseUint(os.Getenv("PORT"), 10, 0)

		if err != nil {
			log.Print(err)
			log.Fatal("Cannot parse port value to uint")
		}

		cfg.Port = port
	}

	if cliArgs.GrpcPort != 0 {
		cfg.GRpcPort = cliArgs.GrpcPort
	} else {
		grpcPort, err := strconv.ParseUint(os.Getenv("GRPC_PORT"), 10, 0)

		if err != nil {
			log.Print(err)
			log.Fatal("Cannot parse grpc port value to uint")
		}

		cfg.GRpcPort = grpcPort
	}

	if cliArgs.StoragePath != "" {
		cfg.FileStoragePath = cliArgs.StoragePath
	} else {
		cfg.FileStoragePath = os.Getenv("STORAGE_PATH")
	}

	cfg.FullAddress = fmt.Sprintf("%s:%d", cfg.IpAddress, cfg.Port)

	cfg.DbConnectionString = os.Getenv("DB_CONNECTION_STRING")

	return cfg
}
