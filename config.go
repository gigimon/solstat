package main

import "flag"

type Conf struct {
	SolanaUrl  string
	NumThreads int
	NumBlocks  int
}

func initializeOptParser() Conf {
	solanaUrl := flag.String("solana-url", "https://api.devnet.solana.com", "Solana RPC URL")
	numThreads := flag.Int("num-threads", 8, "Number of threads to use")
	numBlocks := flag.Int("num-blocks", 16, "Number of blocks to process")

	flag.Parse()
	conf := Conf{*solanaUrl, *numThreads, *numBlocks}
	return conf
}
