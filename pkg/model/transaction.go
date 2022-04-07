package model

import "gorm.io/gorm"

type Transaction struct {
	gorm.Model
	Hash    string `gorm:"uniqueIndex"`
	BlockId uint
}
