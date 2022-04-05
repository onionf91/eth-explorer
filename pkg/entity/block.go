package entity

type BlockHeader struct {
	BlockNumber uint64 `json:"block_number"`
	BlockHash   string `json:"block_hash"`
	BlockTime   uint64 `json:"block_time"`
	ParentHash  string `json:"parent_hash"`
}

type Block struct {
	BlockHeader
	TxHashList []string `json:"transactions"`
}
