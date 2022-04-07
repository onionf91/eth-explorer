package model

import "gorm.io/gorm"

type Block struct {
	gorm.Model
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
	ParentId        *uint
	Parent          *Block
	ReceiptsRoot    string
	Sha3Uncles      string
	Size            float64
	StateRoot       string
	Timestamp       uint64 `gorm:"index"`
	Transactions    []Transaction
	TransactionRoot string
	NephewId        *uint
	Uncles          []Block `gorm:"foreignKey:NephewId"`
}
