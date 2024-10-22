package main

import (
	"context"
	"log"
	"sync"

	"github.com/gigimon/solstat/pkg/config"
	"github.com/gigimon/solstat/pkg/database"
	"github.com/gigimon/solstat/pkg/processor"
	"github.com/gigimon/solstat/pkg/solclient"
)

func main() {
	cfg := config.GetServerConfig()
	solClient := solclient.GetSolanaClient(cfg.Network.SolanaUrl, 30)
	log.Println("Solana client initialized")
	mongoClient := database.InitializeDatabase(cfg.Database.MongoUri, cfg.Database.DBName)
	defer func() {
		if err := mongoClient.Client().Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	log.Println("Mongo client initialized")
	stopCtx, stopCancel := context.WithCancel(context.Background())
	defer stopCancel()

	blockNumbersChan := make(chan uint64, 64)
	blocksChan := make(chan processor.BlockResp, 64)

	log.Println("Starting block worker")
	go processor.RetrieveBlockNumbers(stopCtx, cfg.Network.StartFrom, solClient, blockNumbersChan)

	wg := &sync.WaitGroup{}
	wg.Add(cfg.Network.NumThreads)

	log.Println("Starting block workers to fetch blocks ", cfg.Network.NumThreads)
	for i := 0; i < cfg.Network.NumThreads; i++ {
		go processor.NewBlockWorker(stopCtx, solClient, blockNumbersChan, blocksChan, wg)
	}

	log.Println("Starting processor workers to process blocks ", cfg.Cmd.ProccesorThreads)
	wg.Add(cfg.Cmd.ProccesorThreads)
	for i := 0; i < cfg.Cmd.ProccesorThreads; i++ {
		go processor.ProcessBlock(stopCtx, blocksChan, mongoClient, wg)
	}
	wg.Wait()
}
