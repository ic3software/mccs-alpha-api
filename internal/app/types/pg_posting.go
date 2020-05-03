package types

import (
	"github.com/jinzhu/gorm"
)

type Posting struct {
	gorm.Model
	AccountNumber string  `gorm:"varchar(16);not null;default:''"`
	JournalID     uint    `gorm:"not null"`
	Amount        float64 `gorm:"not null"`
}
