package processor

import (
	"log"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/common"
	"github.com/near/borsh-go"
)

type ComputeBudgetPriceDecoder struct {
	Instruction      uint8
	ComputeUnitPrice uint64
}

type ComputeBudgetUnitLimitDecoder struct {
	Instruction       uint8
	ComputeUnitsLimit uint32
}

type ComputeBudgetData struct {
	ComputeUnitPrice  uint64
	ComputeUnitsLimit uint32
}

func GetComputeBudgetData(tx *client.BlockTransaction) ComputeBudgetData {
	cbIndex := -1
	for index, addr := range tx.Transaction.Message.Accounts {
		if addr == common.ComputeBudgetProgramID {
			cbIndex = index
			break
		}
	}

	cbData := ComputeBudgetData{0, 0}

	for _, i := range tx.Transaction.Message.Instructions {
		if i.ProgramIDIndex == cbIndex && i.Data[0] == 2 {
			decodedData := new(ComputeBudgetUnitLimitDecoder)
			err := borsh.Deserialize(decodedData, i.Data)
			if err != nil {
				log.Println("Failed to decode ComputeBudget data: ", err)
				cbData.ComputeUnitsLimit = 0
			}
			cbData.ComputeUnitsLimit = decodedData.ComputeUnitsLimit
		} else if i.ProgramIDIndex == cbIndex && i.Data[0] == 3 {
			decodedData := new(ComputeBudgetPriceDecoder)
			err := borsh.Deserialize(decodedData, i.Data)
			if err != nil {
				log.Println("Failed to decode ComputeBudget data: ", err)
				cbData.ComputeUnitPrice = 0
			}
			cbData.ComputeUnitPrice = decodedData.ComputeUnitPrice
		}
	}
	return cbData
}
