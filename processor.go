package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/blocto/solana-go-sdk/client"
)

type BlockInfo struct {
	Slot uint64

	CU uint64

	NumTxs uint64

	MinFee uint64
	AvgFee uint64
	MaxFee uint64
}

func ProcessBlock(solClient *client.Client, slotChan <-chan uint64, resultChan chan<- BlockInfo, wg *sync.WaitGroup) {
	// return: block number, a list with transactions (status, fee, cu), used cu
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for slotNumber := range slotChan {
		log.Println("Processing block: ", slotNumber)
		b := BlockInfo{Slot: slotNumber, CU: 0, NumTxs: 0, MinFee: 0, AvgFee: 0, MaxFee: 0}
		block, err := solClient.GetBlock(ctx, slotNumber)
		if err != nil {
			log.Println("Failed to get block: ", err)
			continue
		}
		b.NumTxs = uint64(len(block.Transactions))
		for _, tx := range block.Transactions {
			b.CU += *tx.Meta.ComputeUnitsConsumed
			if b.MinFee == 0 || tx.Meta.Fee < b.MinFee {
				b.MinFee = tx.Meta.Fee
			}
			if tx.Meta.Fee > b.MaxFee {
				b.MaxFee = tx.Meta.Fee
			}
			b.AvgFee += tx.Meta.Fee
		}
		if b.NumTxs > 0 {
			b.AvgFee /= b.NumTxs
		}
		resultChan <- b
	}
	wg.Done()
}
