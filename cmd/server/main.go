package main

import (
	"context"
	"sync"

	"github.com/gigimon/solstat/pkg/config"
	"github.com/gigimon/solstat/pkg/processor"
	"github.com/gigimon/solstat/pkg/solclient"
)

func main() {
	cfg := config.GetServerConfig()
	solClient := solclient.GetSolanaClient(cfg.Network.SolanaUrl, 30)

	stopCtx, stopCancel := context.WithCancel(context.Background())
	defer stopCancel()

	blockNumbersChan := make(chan uint64, 64)
	blocksChan := make(chan processor.BlockResp, 64)
	parsedBlocksChan := make(chan processor.ParsedBlock, len(blocksChan)*2)

	go processor.RetrieveBlockNumbers(stopCtx, cfg.Network.StartFrom, solClient, blockNumbersChan)

	wg := &sync.WaitGroup{}
	wg.Add(cfg.Network.NumThreads)

	for i := 0; i < cfg.Network.NumThreads; i++ {
		go processor.NewBlockWorker(stopCtx, solClient, blockNumbersChan, blocksChan, wg)
	}
	go processor.ProcessBlock(stopCtx, blocksChan, parsedBlocksChan)
}
