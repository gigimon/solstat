package config

import "flag"

type Config struct {
	Network NetworkOptions
	Cmd     CmdOptions
}

type NetworkOptions struct {
	SolanaUrl  string
	NumThreads int
}

type CmdOptions struct {
	NumBlocks int
}

func GetCliConfig() Config {
	solanaUrl := flag.String("solana-url", "https://api.devnet.solana.com", "Solana RPC URL")
	numThreads := flag.Int("num-threads", 8, "Number of threads to use")
	numBlocks := flag.Int("num-blocks", 16, "Number of blocks to process")

	flag.Parse()

	cmdOpts := CmdOptions{*numBlocks}
	networkConf := NetworkOptions{*solanaUrl, *numThreads}
	conf := Config{networkConf, cmdOpts}

	return conf
}

func GetServerConfig() Config {
	solanaUrl := flag.String("solana-url", "https://api.devnet.solana.com", "Solana RPC URL")
	numThreads := flag.Int("num-threads", 8, "Number of threads to use")

	flag.Parse()

	networkConf := NetworkOptions{*solanaUrl, *numThreads}
	conf := Config{}
	conf.Network = networkConf

	return conf
}
