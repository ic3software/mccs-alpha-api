package api

type httpError struct {
	Message string `json:"message"`
}

type httpErrors struct {
	Errors []httpError `json:"errors"`
}
