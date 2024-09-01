package processor

import (
	"context"
	"time"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/types"
)

type ParsedBlock struct {
	Slot uint64

	CU uint64

	NumTxs uint64
	CBTxs  uint64

	Transactions []ParsedTransaction
}

type ParsedTransaction struct {
	Slot       uint64
	Hash       types.Signature
	Signatures []types.Signature
}

func RetrieveBlockNumbers(ctx context.Context, startFrom string, solClient *client.Client, blockChan chan uint64) {
	solCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	startSlot := uint64(0)
	lastIndexedSlot := uint64(0)

	latest, err := solClient.GetSlot(solCtx)
	if err != nil {
		panic(err)
	}
	startSlot = latest

	if startFrom == "continue" {
		lastIndexedSlot = 0 //get this from database
	} else if startFrom == "latest" {
		lastIndexedSlot = startSlot
	} else {
		panic("Invalid startFrom value")
	}

	lastIndexedSlot = startSlot // get this from db and remove
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// get the latest slot
			latest, err := solClient.GetSlot(solCtx)
			if err != nil {
				panic(err)
			}

			// get actual slots numbers
			fromBlock := lastIndexedSlot
			toBlock := uint64(0)

			if latest-lastIndexedSlot < 64 {
				toBlock = latest
			} else {
				toBlock = lastIndexedSlot + 64
			}

			blocks, err := solClient.RpcClient.GetBlocksWithLimit(solCtx, fromBlock, toBlock)
			if err != nil {
				panic(err)
			}

			for _, block := range blocks.Result {
				blockChan <- block
				lastIndexedSlot = block
			}

			// sleep for 1 second
			time.Sleep(1 * time.Second)
		}
	}
}

func ProcessBlock(ctx context.Context, blocksChan chan BlockResp, processedBlocksChan chan ParsedBlock) {
	for {
		select {
		case <-ctx.Done():
			return
		case block := <-blocksChan:
			b := ParsedBlock{}

			b.Slot = block.Slot
			// b.Transactions = block.Block.Transactions
			b.NumTxs = uint64(len(block.Block.Transactions))

			processedBlocksChan <- b
		}
	}
}
