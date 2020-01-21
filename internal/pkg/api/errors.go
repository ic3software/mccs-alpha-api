package api

// HTTPError happens when there's an error for the inputs or operations.
type HTTPError struct {
	Message string `json:"message"`
}
