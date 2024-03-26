package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/near/borsh-go"
)

type ComputeBudgetData struct {
	Instruction  uint8
	ComputeUnits uint64
}

type BlockInfo struct {
	Slot uint64

	CU uint64

	NumTxs uint64
	CBTxs  uint64

	MinFeeSol uint64
	AvgFeeSol uint64
	MaxFeeSol uint64

	MinFee uint64
	AvgFee uint64
	MaxFee uint64
}

func getComputeBudgetPrice(tx *client.BlockTransaction) uint64 {
	cbIndex := -1
	for index, addr := range tx.Transaction.Message.Accounts {
		if addr.String() == "ComputeBudget111111111111111111111111111111" {
			cbIndex = index
			break
		}
	}
	for _, i := range tx.Transaction.Message.Instructions {
		if i.ProgramIDIndex == cbIndex && i.Data[0] == 3 {
			decodedData := new(ComputeBudgetData)
			err := borsh.Deserialize(decodedData, i.Data)
			if err != nil {
				log.Println("Failed to decode ComputeBudget data: ", err)
				return 0
			}
			return decodedData.ComputeUnits
		}
	}
	return 0
}

func ProcessBlock(solClient *client.Client, slotChan <-chan uint64, resultChan chan<- BlockInfo, wg *sync.WaitGroup) {
	// return: block number, a list with transactions (status, fee, cu), used cu
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for slotNumber := range slotChan {
		log.Println("Processing block: ", slotNumber)
		b := BlockInfo{Slot: slotNumber, CU: 0, NumTxs: 0, MinFeeSol: 0, AvgFeeSol: 0, MaxFeeSol: 0}
		block, err := solClient.GetBlock(ctx, slotNumber)
		if err != nil {
			log.Println("Failed to get block: ", err)
			continue
		}
		b.NumTxs = uint64(len(block.Transactions))
		for _, tx := range block.Transactions {
			b.CU += *tx.Meta.ComputeUnitsConsumed

			if b.MinFeeSol == 0 || tx.Meta.Fee < b.MinFeeSol {
				b.MinFeeSol = tx.Meta.Fee
			}
			if tx.Meta.Fee > b.MaxFeeSol {
				b.MaxFeeSol = tx.Meta.Fee
			}
			b.AvgFeeSol += tx.Meta.Fee

			cuPrice := getComputeBudgetPrice(&tx)
			if cuPrice != 0 {
				b.CBTxs++
				if b.MinFee == 0 || cuPrice < b.MinFee {
					b.MinFee = cuPrice
				}
				if cuPrice > b.MaxFee {
					b.MaxFee = cuPrice
				}
				b.AvgFee += cuPrice
			}
		}
		if b.NumTxs > 0 {
			b.AvgFeeSol /= b.NumTxs
		}

		if b.CBTxs > 0 {
			b.AvgFee /= b.CBTxs
		}

		resultChan <- b
	}
	wg.Done()
}
