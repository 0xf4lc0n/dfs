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
	IpAddress   string
	Port        uint64
	FullAddress string
}

func Create() *Config {
	var cliArgs CliArgs
	kong.Parse(&cliArgs)

	err := godotenv.Load(".env")

	if err != nil {
		log.Printf("Cannot load .env file")
	}

	var ipAddress string

	if cliArgs.IpAddress != "" {
		ipAddress = cliArgs.IpAddress
	} else {
		ipAddress = os.Getenv("IP_ADDRESS")
	}

	var port uint64

	if cliArgs.Port != 0 {
		port = cliArgs.Port
	} else {
		var err error
		port, err = strconv.ParseUint(os.Getenv("PORT"), 10, 0)

		if err != nil {
			log.Print(err)
			log.Fatal("Cannot parse port value to uint")
		}
	}

	fullAddress := fmt.Sprintf("%s:%d", ipAddress, port)

	cfg := &Config{
		IpAddress:   ipAddress,
		Port:        port,
		FullAddress: fullAddress,
	}

	return cfg
}
