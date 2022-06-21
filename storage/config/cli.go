package config

type CliArgs struct {
	IpAddress   string `help:"Ip address"`
	Port        uint64 `help:"Network port"`
	GrpcPort    uint64 `help:"Network port for GRPC server"`
	StoragePath string `help:"Path where all files will be stored"`
}
