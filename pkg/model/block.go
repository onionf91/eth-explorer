package model

import "gorm.io/gorm"

type Block struct {
	gorm.Model
	ID              int
	Difficulty      uint64
	ExtraData       string
	GasLimit        uint64
	GasUsed         uint64
	Hash            string
	LogsBloom       string
	Miner           string
	MixHash         string
	Nonce           uint64
	Number          uint64
	ParentHash      *Block `gorm:"foreignKey:Hash"`
	ReceiptsRoot    string
	Sha3Uncles      string
	Size            float64
	StateRoot       string
	Timestamp       uint64
	TotalDifficulty uint64
	Transactions    []*Transaction `gorm:"foreignKey:Hash"`
	TransactionRoot string
	Uncles          []*Block `gorm:"foreignKey:Hash"`
}
