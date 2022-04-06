package model

type Transaction struct {
	Id   uint64 `gorm:"primary_key;auto_increment;not_null"`
	Hash string `gorm:"uniqueIndex"`
}
