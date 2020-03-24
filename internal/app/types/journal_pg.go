package types

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Journal struct {
	gorm.Model
	// Journal has many postings, JournalID is the foreign key
	Postings []Posting

	TransferID string `gorm:"type:varchar(27);not null;default:''"`

	InitiatedBy string `gorm:"varchar(16);not null;default:''"`

	FromAccountNumber string `gorm:"varchar(16);not null;default:''"`
	FromEmail         string `gorm:"type:varchar(120);not null;default:''"`
	FromEntityName    string `gorm:"type:varchar(120);not null;default:''"`

	ToAccountNumber string `gorm:"varchar(16);not null;default:''"`
	ToEmail         string `gorm:"type:varchar(120);not null;default:''"`
	ToEntityName    string `gorm:"type:varchar(120);not null;default:''"`

	Amount      float64 `gorm:"not null;default:0"`
	Description string  `gorm:"type:varchar(510);not null;default:''"`
	Type        string  `gorm:"type:varchar(31);not null;default:'transfer'"`
	Status      string  `gorm:"type:varchar(31);not null;default:''"`

	CompletedAt time.Time

	CancellationReason string `gorm:"type:varchar(510);not null;default:''"`
}
