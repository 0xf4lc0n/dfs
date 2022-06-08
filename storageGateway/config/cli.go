package config

type CliArgs struct {
	IpAddress string `help:"Ip address"`
	Port      uint64 `help:"Network port"`
}
