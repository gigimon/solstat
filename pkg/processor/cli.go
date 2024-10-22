package processor

type BlockCUStat struct {
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

func ProcessBlockCLI(block BlockResp) BlockCUStat {
	b := BlockCUStat{Slot: block.Slot, CU: 0, NumTxs: 0, MinFeeSol: 0, AvgFeeSol: 0, MaxFeeSol: 0}

	b.NumTxs = uint64(len(block.Block.Transactions))
	for _, tx := range block.Block.Transactions {
		b.CU += *tx.Meta.ComputeUnitsConsumed

		if b.MinFeeSol == 0 || tx.Meta.Fee < b.MinFeeSol {
			b.MinFeeSol = tx.Meta.Fee
		}
		if tx.Meta.Fee > b.MaxFeeSol {
			b.MaxFeeSol = tx.Meta.Fee
		}
		b.AvgFeeSol += tx.Meta.Fee

		cuPrice := GetComputeBudgetData(&tx).ComputeUnitPrice
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
	return b
}
