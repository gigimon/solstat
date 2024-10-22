package processor

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/types"
	"go.mongodb.org/mongo-driver/mongo"
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

type ServerProcessedBlock struct {
	Slot       uint64    `json:"slot"`
	ParentSlot uint64    `json:"parent_slot"`
	SlotTime   time.Time `json:"slot_time"`

	NumTxs        uint64 `json:"num_txs"`         // Number of transactions in the block
	NumSuccessTxs uint64 `json:"num_success_txs"` // Number of successful transactions in the block
	NumCBTxs      uint64 `json:"num_cb_txs"`      // Number of transactions that have compute budget

	MinCUPrice uint64 `json:"min_cu_price"`
	AvgCUPrice uint64 `json:"avg_cu_price"`
	MaxCUPrice uint64 `json:"max_cu_price"`

	MinPriorityFee uint64 `json:"min_priority_fee"` // Minimum priority fee in lamports
	AvgPriorityFee uint64 `json:"avg_priority_fee"` // Average priority fee in lamports
	MaxPriorityFee uint64 `json:"max_priority_fee"` // Maximum priority fee in lamports

	CU       uint64 `json:"cu"`        // Total compute units consumed in the block
	BlockFee uint64 `json:"block_fee"` // Total block fee in lamports

}

func RetrieveBlockNumbers(ctx context.Context, startFrom string, solClient *client.Client, blockChan chan uint64) {
	solCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	startSlot := uint64(0)
	lastIndexedSlot := uint64(0)

	latest, err := solClient.GetSlot(solCtx)
	log.Println("Latest slot: ", latest)
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
			log.Println("Stopping block number retrieval by context Done")
			return
		default:
			// get the latest slot
			solCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			latest, err := solClient.GetSlot(solCtx)
			if err != nil {
				log.Println("Failed to get latest slot: ", err)
			}

			// get actual slots numbers
			fromBlock := lastIndexedSlot
			blocksLimit := uint64(0)

			if latest-lastIndexedSlot < 64 {
				blocksLimit = latest - lastIndexedSlot
			} else {
				blocksLimit = 64
			}

			log.Println("Getting blocks from ", fromBlock, " to ", blocksLimit)
			blocks, err := solClient.RpcClient.GetBlocksWithLimit(solCtx, fromBlock, blocksLimit)
			if err != nil {
				log.Println("Failed to get blocks: ", err)
			}
			log.Println("Fetched blocks: ", blocks.Result)

			for _, block := range blocks.Result {
				log.Println("Send to processing block: ", block)
				blockChan <- block
				lastIndexedSlot = block
			}

			// sleep for 1 second
			time.Sleep(1 * time.Second)
		}
	}
}

func ProcessBlock(ctx context.Context, blocksChan chan BlockResp, mongoClient *mongo.Database, wg *sync.WaitGroup) {
	for {
		log.Println("Blocks in queue: ", len(blocksChan))
		select {
		case <-ctx.Done():
			wg.Done()
			log.Println("Stopping block processing by context Done")
			return
		case block := <-blocksChan:
			log.Println("Processing block: ", block.Slot)
			b := ServerProcessedBlock{}

			b.Slot = block.Slot
			b.ParentSlot = block.Block.ParentSlot
			b.SlotTime = *block.Block.BlockTime

			b.NumTxs = uint64(len(block.Block.Transactions))
			b.NumSuccessTxs = 0
			b.NumCBTxs = 0
			b.CU = 0
			b.BlockFee = 0

			for _, tx := range block.Block.Transactions {
				if tx.Meta.Err == nil {
					b.NumSuccessTxs++
				}

				// get the compute budget data
				cbData := GetComputeBudgetData(&tx)

				if cbData.ComputeUnitsLimit > 0 {
					b.NumCBTxs++
					b.CU += uint64(cbData.ComputeUnitsLimit)

					if cbData.ComputeUnitPrice < b.MinCUPrice {
						b.MinCUPrice = cbData.ComputeUnitPrice
					}
					if cbData.ComputeUnitPrice > b.MaxCUPrice {
						b.MaxCUPrice = cbData.ComputeUnitPrice
					}
					b.AvgCUPrice += cbData.ComputeUnitPrice

					txFee := uint64(cbData.ComputeUnitsLimit) * cbData.ComputeUnitPrice
					b.BlockFee += txFee

					if txFee < b.MinPriorityFee {
						b.MinPriorityFee = txFee
					}
					if txFee > b.MaxPriorityFee {
						b.MaxPriorityFee = txFee
					}
					b.AvgPriorityFee += txFee

				} else {
					b.CU += *tx.Meta.ComputeUnitsConsumed
					b.BlockFee += tx.Meta.Fee
				}

				// get the block fee
				b.BlockFee += tx.Meta.Fee
			}
			if b.NumCBTxs > 0 {
				b.AvgCUPrice /= b.NumCBTxs
				b.AvgPriorityFee /= b.NumCBTxs
			}
			log.Println("Writing block to database: ", b.Slot)
			_, err := mongoClient.Collection("blocks").InsertOne(ctx, b)
			if err != nil {
				log.Println("Failed to write block to database: ", err)
			}
		}
	}
}
