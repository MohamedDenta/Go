package dashboard

import (
	"time"
)

type blockInfo struct {
	BlockId           int
	Blockcreationdate time.Time
	NoOfTransaction   int
	GnodeID           string
}

type dashboard struct {
	Blockinfo   []blockInfo
	Blockheight int
	PayloadBase string
}
type statStruct struct {
	NumberOfValidator    int
	NumberOfBlocks       int
	NumberOfTransactions int
	NumberOfStakeCoin    []string
}
