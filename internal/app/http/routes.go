package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/internal/app/http/controller"
	"github.com/ic3network/mccs-alpha-api/internal/app/http/middleware"
)

func RegisterRoutes(r *mux.Router) {
	public := r.PathPrefix("/api/v1").Subrouter()
	public.Use(middleware.CORS(), middleware.Recover(), middleware.NoCache(), middleware.Logging(), middleware.GetLoggedInUser())
	private := r.PathPrefix("/api/v1").Subrouter()
	private.Use(middleware.CORS(), middleware.Recover(), middleware.NoCache(), middleware.Logging(), middleware.GetLoggedInUser(), middleware.RequireUser())
	adminPublic := r.PathPrefix("/api/v1/admin").Subrouter()
	adminPublic.Use(middleware.CORS(), middleware.Recover(), middleware.NoCache(), middleware.Logging(), middleware.GetLoggedInUser())
	adminPrivate := r.PathPrefix("/api/v1/admin").Subrouter()
	adminPrivate.Use(middleware.CORS(), middleware.Recover(), middleware.NoCache(), middleware.Logging(), middleware.GetLoggedInUser(), middleware.RequireAdmin())

	// Serving static files.
	fs := http.FileServer(http.Dir("web/static"))
	public.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	controller.ServiceDiscovery.RegisterRoutes(public, private)
	controller.EntityHandler.RegisterRoutes(public, private)
	controller.UserHandler.RegisterRoutes(public, private, adminPublic, adminPrivate)
	controller.TransferHandler.RegisterRoutes(public, private)

	controller.AdminEntityHandler.RegisterRoutes(adminPublic, adminPrivate)
	controller.AdminUserHandler.RegisterRoutes(public, private, adminPublic, adminPrivate)
	controller.AdminHistoryHandler.RegisterRoutes(public, private, adminPublic, adminPrivate)
	controller.AdminTagHandler.RegisterRoutes(adminPublic, adminPrivate)
	controller.AdminTransactionHandler.RegisterRoutes(public, private, adminPublic, adminPrivate)
	controller.CategoryHandler.RegisterRoutes(public, private, adminPublic, adminPrivate)
	controller.LogHandler.RegisterRoutes(public, private, adminPublic, adminPrivate)

	controller.AccountHandler.RegisterRoutes(public, private, adminPublic, adminPrivate)
	controller.TagHandler.RegisterRoutes(public, private, adminPublic, adminPrivate)
}
