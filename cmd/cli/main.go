package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/gigimon/solstat/pkg/config"
	"github.com/gigimon/solstat/pkg/processor"
	"github.com/gigimon/solstat/pkg/solclient"
)

func main() {
	opts := config.GetCliConfig()

	log.Println("Create Solana client with URL: ", opts.Network.SolanaUrl)

	solClient := solclient.GetSolanaClient(opts.Network.SolanaUrl, 30)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	currentBlock, err := solClient.GetSlot(ctx)
	if err != nil {
		log.Fatal("Failed to get current block: ", err)
	}

	blocks, err := solClient.RpcClient.GetBlocksWithLimit(ctx, currentBlock-uint64(opts.Cmd.NumBlocks), uint64(opts.Cmd.NumBlocks))
	if err != nil {
		log.Fatal("Failed to get blocks: ", err)
	}
	log.Println("Got blocks numbers: ", blocks.Result)

	slotChan := make(chan uint64, opts.Cmd.NumBlocks)
	resultChan := make(chan processor.BlockResp, opts.Cmd.NumBlocks)

	wg := &sync.WaitGroup{}
	wg.Add(opts.Network.NumThreads)

	for i := 0; i < opts.Cmd.NumBlocks; i++ {
		slotChan <- blocks.Result[i]
	}
	close(slotChan)

	stopCtx, stopCancel := context.WithCancel(context.Background())
	defer stopCancel()

	for i := 0; i < opts.Cmd.NumBlocks; i++ {
		go processor.NewBlockWorker(stopCtx, solClient, slotChan, resultChan, wg)
	}
	wg.Wait()
	log.Println("All blocks processed")

	if len(resultChan) != opts.Cmd.NumBlocks {
		log.Println("Not all blocks are processed, only ", len(resultChan), " blocks are processed")
	}

	if len(resultChan) == 0 {
		log.Fatal("No block is processed")
	}

	results := make([]processor.BlockCUStat, opts.Cmd.NumBlocks)
	avgFee := uint64(0)
	cbAvgFee := uint64(0)

	for i := 0; i < len(resultChan); i++ {
		block := <-resultChan
		results[i] = processor.ProcessBlockCLI(block)
		avgFee += results[i].AvgFeeSol
		cbAvgFee += results[i].AvgFee
	}

	avgFee = avgFee / uint64(opts.Cmd.NumBlocks)
	cbAvgFee = cbAvgFee / uint64(opts.Cmd.NumBlocks)

	sort.Slice(results, func(i, j int) bool {
		return results[i].Slot < results[j].Slot
	})

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Block", "CU", "Min fee (SOL)", "Max fee (SOL)", "Avg fee (SOL)", "CB Min (lam)", "CB Max (lam)", "CB Avg (lam)", "Num txns"})
	for _, r := range results {
		t.AppendRow(table.Row{
			r.Slot,
			fmt.Sprintf("%d (%.2f%%)",
				r.CU,
				float64(r.CU)/48000000*100),
			fmt.Sprintf("%f", float64(r.MinFeeSol)/1000000000),
			fmt.Sprintf("%f", float64(r.MaxFeeSol)/1000000000),
			fmt.Sprintf("%f", float64(r.AvgFeeSol)/1000000000),
			fmt.Sprintf("%f", float64(r.MinFee)/1000000),
			fmt.Sprintf("%f", float64(r.MaxFee)/1000000),
			fmt.Sprintf("%f", float64(r.AvgFee)/1000000),
			r.NumTxs})
	}
	t.AppendFooter(table.Row{
		"",
		"",
		"",
		"",
		fmt.Sprintf("%f", float64(avgFee)/1000000000),
		"",
		"",
		fmt.Sprintf("%f", float64(cbAvgFee)/1000000),
	})
	t.Render()
}
