package types

import (
	"github.com/jinzhu/gorm"
)

type BalanceLimit struct {
	gorm.Model
	// `BalanceLimit` belongs to `Account`, `AccountID` is the foreign key
	Account       Account
	AccountNumber string  `json:"accountNumber,omitempty" gorm:"type:varchar(16);not null;unique_index"`
	MaxNegBal     float64 `json:"maxNegBal,omitempty" gorm:"type:real;not null"`
	MaxPosBal     float64 `json:"maxPosBal,omitempty" gorm:"type:real;not null"`
}
