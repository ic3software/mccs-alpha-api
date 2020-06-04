package types

// EntityESRecord is the data that will store into the elastic search.
type EntityESRecord struct {
	ID     string `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	Email  string `json:"email,omitempty"`
	Status string `json:"status,omitempty"`
	// Tags
	Offers     []*TagField `json:"offers,omitempty"`
	Wants      []*TagField `json:"wants,omitempty"`
	Categories []string    `json:"categories,omitempty"`
	// Address
	City    string `json:"city,omitempty"`
	Region  string `json:"region,omitempty"`
	Country string `json:"country,omitempty"`
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
