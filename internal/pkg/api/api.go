package api

import (
	"encoding/json"
	"net/http"

	"github.com/ic3network/mccs-alpha-api/internal/pkg/l"
	"go.uber.org/zap"
)

// Respond return an object with specific status as JSON.
func Respond(w http.ResponseWriter, r *http.Request, status int, potentialData ...interface{}) {
	if len(potentialData) > 0 {
		data := potentialData[0]

		// change error into a real JSON serializable object
		if e, ok := data.(error); ok {
			data = HTTPError{Message: e.Error()}
		}

		js, err := json.Marshal(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			l.Logger.Error("[ERROR] Marshaling data error:", zap.Error(err))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write(js)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
}
