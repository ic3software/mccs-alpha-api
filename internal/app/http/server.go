package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/util/l"
	"go.uber.org/zap"
)

// AppServer contains the information to run a server.
type appServer struct{}

var AppServer = &appServer{}

// Run will start the http server.
func (a *appServer) Run(port string) {
	r := mux.NewRouter().StrictSlash(true)
	RegisterRoutes(r)

	headersOk := handlers.AllowedHeaders([]string{"Authorization", "Content-Type"})

	srv := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%s", port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      handlers.CORS(headersOk)(r),
	}

	l.Logger.Info("app is running at localhost:" + port)

	if err := srv.ListenAndServe(); err != nil {
		l.Logger.Fatal("ListenAndServe failed", zap.Error(err))
	}
}
