package types

import (
	"time"
)

type Transfer struct {
	ID             uint // Journal ID
	TransactionID  string
	IsInitiator    bool
	InitiatedBy    uint
	FromID         uint
	FromEmail      string
	FromEntityName string
	ToID           uint
	ToEmail        string
	ToEntityName   string
	Amount         float64
	Description    string
	Status         string
	CreatedAt      time.Time
}
