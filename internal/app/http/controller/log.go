package controller

import (
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

type logHandler struct {
	once *sync.Once
}

var LogHandler = newLogHandler()

func newLogHandler() *logHandler {
	return &logHandler{
		once: new(sync.Once),
	}
}

func (lh *logHandler) RegisterRoutes(
	public *mux.Router,
	private *mux.Router,
	adminPublic *mux.Router,
	adminPrivate *mux.Router,
) {
	lh.once.Do(func() {
		adminPrivate.Path("/log/search").HandlerFunc(lh.searchLog()).Methods("GET")
	})
}

func (lh *logHandler) searchLog() func(http.ResponseWriter, *http.Request) {
	// t := template.NewView("/admin/log")
	// type formData struct {
	// 	Email    string
	// 	DateFrom string
	// 	DateTo   string
	// 	Category string
	// 	Page     int
	// }
	// type response struct {
	// 	FormData    formData
	// 	UserActions []*types.UserAction
	// 	TotalPages  int
	// }
	return func(w http.ResponseWriter, r *http.Request) {
		// 	q := r.URL.Query()

		// 	page, err := strconv.Atoi(q.Get("page"))
		// 	if err != nil {
		// 		l.Logger.Error("SearchUserLogs failed", zap.Error(err))
		// 		t.Error(w, r, nil, err)
		// 		return
		// 	}

		// 	f := formData{
		// 		Email:    q.Get("email"),
		// 		Category: q.Get("category"),
		// 		DateFrom: q.Get("date-from"),
		// 		DateTo:   q.Get("date-to"),
		// 		Page:     page,
		// 	}
		// 	res := response{FormData: f}

		// 	c := types.UserActionSearchCriteria{
		// 		Email:    f.Email,
		// 		Category: f.Category,
		// 		DateFrom: util.ParseTime(f.DateFrom),
		// 		DateTo:   util.ParseTime(f.DateTo),
		// 	}

		// 	userAction, totalPages, err := logic.UserAction.Find(&c, int64(f.Page))
		// 	res.TotalPages = totalPages
		// 	res.UserActions = userAction
		// 	if err != nil {
		// 		l.Logger.Error("SearchUserLogs failed", zap.Error(err))
		// 		t.Error(w, r, res, err)
		// 		return
		// 	}

		// 	t.Render(w, r, res, nil)
	}
}
