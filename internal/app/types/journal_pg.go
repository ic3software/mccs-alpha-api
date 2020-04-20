package types

import (
	"time"

	"github.com/ic3network/mccs-alpha-api/global/constant"
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

func JournalsToTransfers(journals []*Journal, queryingAccountNumber string) []*TransferRespond {
	transfers := []*TransferRespond{}

	for _, j := range journals {
		t := &TransferRespond{
			TransferID:  j.TransferID,
			Description: j.Description,
			Amount:      j.Amount,
			CreatedAt:   &j.CreatedAt,
			Status:      j.Status,
		}
		if j.InitiatedBy == queryingAccountNumber {
			t.IsInitiator = true
		}
		if j.FromAccountNumber == queryingAccountNumber {
			t.Transfer = "out"
			t.AccountNumber = j.ToAccountNumber
			t.EntityName = j.ToEntityName
		} else {
			t.Transfer = "in"
			t.AccountNumber = j.FromAccountNumber
			t.EntityName = j.FromEntityName
		}
		if j.Status == constant.Transfer.Completed {
			t.CompletedAt = &j.UpdatedAt
		}

		transfers = append(transfers, t)
	}

	return transfers
}
