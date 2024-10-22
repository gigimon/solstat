package processor

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/blocto/solana-go-sdk/client"
)

type BlockResp struct {
	Slot  uint64
	Block *client.Block
}

func NewBlockWorker(ctx context.Context, solClient *client.Client, slotChan <-chan uint64, resultChan chan<- BlockResp, wg *sync.WaitGroup) {
	solCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			log.Println("Context done in block worker")
			wg.Done()
			return
		case slot, ok := <-slotChan:
			log.Println("Blocks numbers in queue: ", len(slotChan))
			if !ok {
				log.Println("Slot channel closed")
				wg.Done()
				return
			}
			block, err := solClient.GetBlock(solCtx, slot)
			if err != nil {
				log.Println("Failed to get block: ", err)
				continue
			}
			log.Println("Fetched block: ", block.Blockhash)
			log.Println("Blocks in block result queue: ", len(resultChan))
			resultChan <- BlockResp{Slot: slot, Block: block}
		}
	}
}
