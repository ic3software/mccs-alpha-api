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
}

type ESFindEntityResult struct {
	IDs             []string
	NumberOfResults int
	TotalPages      int
}
