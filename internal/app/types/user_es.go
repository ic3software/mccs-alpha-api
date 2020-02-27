package types

// UserESRecord is the data that will store into the elastic search.
type UserESRecord struct {
	UserID    string `json:"userID"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Email     string `json:"email,omitempty"`
}
