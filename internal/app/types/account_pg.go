package types

import (
	"github.com/jinzhu/gorm"
)

type Account struct {
	gorm.Model
	// Account has many postings, AccountID is the foreign key
	Postings      []Posting
	AccountNumber string  `gorm:"type:varchar(16);not null;unique_index"`
	Balance       float64 `gorm:"not null;default:0"`
}
