package api

import (
	"encoding/json"
	"log"
	"net/http"
)

// Respond return an object with specific status as JSON.
func Respond(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	// change error into a real JSON serializable object
	if e, ok := data.(error); ok {
		data = HTTPError{Message: e.Error()}
	}

	js, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("[ERROR] Marshaling data error: %+v \t data: %+v \n", err.Error(), data)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
}