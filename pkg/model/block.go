package model

type Block struct {
	Id              uint64 `gorm:"primary_key;auto_increment;not_null"`
	Difficulty      uint64
	ExtraData       string
	GasLimit        uint64
	GasUsed         uint64
	Hash            string `gorm:"uniqueIndex"`
	LogsBloom       string
	Miner           string
	MixHash         string
	Nonce           uint64
	Number          uint64 `gorm:"uniqueIndex"`
	ParentHash      *Block `gorm:"foreignKey:Hash"`
	ReceiptsRoot    string
	Sha3Uncles      string
	Size            float64
	StateRoot       string
	Timestamp       uint64 `gorm:"index"`
	TotalDifficulty uint64
	Transactions    []*Transaction `gorm:"foreignKey:Hash"`
	TransactionRoot string
	Uncles          []*Block `gorm:"foreignKey:Hash"`
}
