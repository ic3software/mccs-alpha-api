package types

import "time"

type JournalESRecord struct {
	TransferID        string    `json:"transferID,omitempty"`
	FromAccountNumber string    `json:"fromAccountNumber,omitempty"`
	ToAccountNumber   string    `json:"toAccountNumber,omitempty"`
	Status            string    `json:"status,omitempty"`
	CreatedAt         time.Time `json:"createdAt,omitempty"`
}

type ESFindJournalResult struct {
	IDs             []string
	NumberOfResults int
	TotalPages      int
}
