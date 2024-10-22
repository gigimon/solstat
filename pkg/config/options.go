package config

import "flag"

type Config struct {
	Network  NetworkOptions
	Cmd      CmdOptions
	Database DatabaseOptions
}

type NetworkOptions struct {
	SolanaUrl  string
	NumThreads int
	StartFrom  string
}

type CmdOptions struct {
	NumBlocks        int
	ProccesorThreads int
}

type DatabaseOptions struct {
	MongoUri string
	DBName   string
}

func GetCliConfig() Config {
	solanaUrl := flag.String("solana-url", "https://api.devnet.solana.com", "Solana RPC URL")
	numThreads := flag.Int("num-threads", 8, "Number of threads to use")
	numBlocks := flag.Int("num-blocks", 16, "Number of blocks to process")

	flag.Parse()

	cmdOpts := CmdOptions{}
	cmdOpts.NumBlocks = *numBlocks
	networkConf := NetworkOptions{*solanaUrl, *numThreads, ""}
	conf := Config{}
	conf.Network = networkConf
	conf.Cmd = cmdOpts

	return conf
}

func GetServerConfig() Config {
	solanaUrl := flag.String("solana-url", "https://api.devnet.solana.com", "Solana RPC URL")
	numThreads := flag.Int("num-threads", 8, "Number of threads to use")
	startFrom := flag.String("start-from", "latest", "Start from block number")
	proccesorThreads := flag.Int("processor-threads", 8, "Number of threads to use for processing")
	uri := flag.String("mongo-uri", "localhost", "MongoDB URI")
	dbname := flag.String("mongo-db", "solstat", "Database name")

	flag.Parse()

	conf := Config{}
	conf.Network = NetworkOptions{*solanaUrl, *numThreads, *startFrom}
	conf.Cmd.ProccesorThreads = *proccesorThreads
	conf.Database = DatabaseOptions{*uri, *dbname}

	return conf
}

func GetHttpConfig() Config {
	conf := Config{}
	uri := flag.String("mongo-uri", "localhost", "MongoDB URI")
	dbname := flag.String("mongo-db", "solstat", "Database name")
	conf.Database = DatabaseOptions{*uri, *dbname}

	return conf
}
