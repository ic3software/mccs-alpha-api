package types

// EntityESRecord is the data that will store into the elastic search.
type EntityESRecord struct {
	EntityID        string      `json:"entityID,omitempty"`
	EntityName      string      `json:"entityName,omitempty"`
	Offers          []*TagField `json:"offers,omitempty"`
	Wants           []*TagField `json:"wants,omitempty"`
	LocationCity    string      `json:"locationCity,omitempty"`
	LocationCountry string      `json:"locationCountry,omitempty"`
	Status          string      `json:"status,omitempty"`
	Categories      []string    `json:"categories,omitempty"`
	// Account
	AccountNumber string   `json:"accountNumber,omitempty"`
	Balance       *float64 `json:"balance,omitempty"`
	MaxNegBal     *float64 `json:"maxNegBal,omitempty"`
	MaxPosBal     *float64 `json:"maxPosBal,omitempty"`
}

type ESSearchEntityResult struct {
	IDs             []string
	NumberOfResults int
	TotalPages      int
}
