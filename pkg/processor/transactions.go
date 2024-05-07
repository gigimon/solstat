package processor

import (
	"log"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/common"
	"github.com/near/borsh-go"
)

type ComputeBudgetData struct {
	Instruction  uint8
	ComputeUnits uint64
}

func GetComputeBudgetPrice(tx *client.BlockTransaction) uint64 {
	cbIndex := -1
	for index, addr := range tx.Transaction.Message.Accounts {
		if addr == common.ComputeBudgetProgramID {
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
