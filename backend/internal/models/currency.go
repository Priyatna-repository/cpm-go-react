package models

type Currency struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Name     string `gorm:"size:120;not null" json:"name"`
	Code     string `gorm:"size:3;uniqueIndex;not null" json:"code"`
	Symbol   string `gorm:"size:5;not null" json:"symbol"`
	Decimals int16  `gorm:"not null;default:2" json:"decimals"`
}
