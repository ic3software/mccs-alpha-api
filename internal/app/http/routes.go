package http

import (
	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/internal/app/http/controller"
	"github.com/ic3network/mccs-alpha-api/internal/app/http/middleware"
)

func RegisterRoutes(r *mux.Router) {
	public := r.PathPrefix("/api/v1").Subrouter()
	public.Use(middleware.Recover(), middleware.NoCache(), middleware.Logging(), middleware.GetLoggedInUser())
	private := r.PathPrefix("/api/v1").Subrouter()
	private.Use(middleware.Recover(), middleware.NoCache(), middleware.Logging(), middleware.GetLoggedInUser(), middleware.RequireUser())
	adminPublic := r.PathPrefix("/api/v1/admin").Subrouter()
	adminPublic.Use(middleware.Recover(), middleware.NoCache(), middleware.Logging(), middleware.GetLoggedInUser())
	adminPrivate := r.PathPrefix("/api/v1/admin").Subrouter()
	adminPrivate.Use(middleware.Recover(), middleware.NoCache(), middleware.Logging(), middleware.GetLoggedInUser(), middleware.RequireAdmin())

	controller.ServiceDiscovery.RegisterRoutes(public, private)
	controller.UserHandler.RegisterRoutes(public, private, adminPublic, adminPrivate)
	controller.AdminUserHandler.RegisterRoutes(public, private, adminPublic, adminPrivate)
	controller.EntityHandler.RegisterRoutes(public, private, adminPublic, adminPrivate)
	controller.TagHandler.RegisterRoutes(public, private, adminPublic, adminPrivate)
	controller.CategoryHandler.RegisterRoutes(public, private, adminPublic, adminPrivate)
	controller.TransferHandler.RegisterRoutes(public, private, adminPublic, adminPrivate)
	controller.UserAction.RegisterRoutes(adminPrivate)
}
