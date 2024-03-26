package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/rpc"
	"github.com/jedib0t/go-pretty/v6/table"
)

func main() {
	opts := initializeOptParser()

	log.Println("Create Solana client with URL: ", opts.SolanaUrl)

	httpTransport := http.Transport{
		ReadBufferSize: 128000,
	}
	httpClient := &http.Client{Transport: &httpTransport, Timeout: 30 * time.Second}

	solClient := client.New(rpc.WithEndpoint(opts.SolanaUrl), rpc.WithHTTPClient(httpClient))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	currentBlock, err := solClient.GetSlot(ctx)
	if err != nil {
		log.Fatal("Failed to get current block: ", err)
	}

	blocks, err := solClient.RpcClient.GetBlocksWithLimit(ctx, currentBlock-uint64(opts.NumBlocks), uint64(opts.NumBlocks))
	if err != nil {
		log.Fatal("Failed to get blocks: ", err)
	}
	log.Println("Got blocks numbers: ", blocks.Result)

	slotChan := make(chan uint64, opts.NumBlocks)
	resultChan := make(chan BlockInfo, opts.NumBlocks)

	wg := &sync.WaitGroup{}
	wg.Add(opts.NumThreads)

	for i := 0; i < opts.NumBlocks; i++ {
		slotChan <- blocks.Result[i]
	}
	close(slotChan)

	for i := 0; i < opts.NumThreads; i++ {
		go ProcessBlock(solClient, slotChan, resultChan, wg)
	}
	wg.Wait()
	log.Println("All blocks processed")

	if len(resultChan) != opts.NumBlocks {
		log.Println("Not all blocks are processed, only ", len(resultChan), " blocks are processed")
	}

	if len(resultChan) == 0 {
		log.Fatal("No block is processed")
	}

	results := make([]BlockInfo, opts.NumBlocks)
	avgFee := uint64(0)

	channelCount := len(resultChan)

	for i := 0; i < channelCount; i++ {
		results[i] = <-resultChan
		avgFee += results[i].AvgFee
	}
	avgFee = avgFee / uint64(opts.NumBlocks)

	sort.Slice(results, func(i, j int) bool {
		return results[i].Slot < results[j].Slot
	})

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Block", "CU", "Min fee", "Max fee", "Avg fee", "Num txns"})
	for _, r := range results {
		t.AppendRow(table.Row{r.Slot, fmt.Sprintf("%d (%.2f%%)", r.CU, float64(r.CU)/48000000*100), r.MinFee, r.MaxFee, r.AvgFee, r.NumTxs})
	}
	t.AppendFooter(table.Row{"", "", "", "", avgFee, ""})
	t.Render()
}
