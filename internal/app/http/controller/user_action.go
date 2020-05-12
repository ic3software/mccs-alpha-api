package controller

import (
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/internal/app/api"
	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/util/l"
	"go.uber.org/zap"
)

var UserAction = newUserAction()

type userAction struct {
	once *sync.Once
}

func newUserAction() *userAction {
	return &userAction{
		once: new(sync.Once),
	}
}

func (ua *userAction) RegisterRoutes(adminPrivate *mux.Router) {
	ua.once.Do(func() {
		adminPrivate.Path("/logs").HandlerFunc(ua.search()).Methods("GET")
	})
}

func (ua *userAction) search() func(http.ResponseWriter, *http.Request) {
	type meta struct {
		NumberOfResults int `json:"numberOfResults"`
		TotalPages      int `json:"totalPages"`
	}
	type respond struct {
		Data []*types.UserActionESRecord `json:"data"`
		Meta meta                        `json:"meta"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := types.NewAdminSearchLog(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		found, err := logic.UserAction.Search(req)
		if err != nil {
			l.Logger.Error("[Error] TransferHandler.searchTransfers failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		api.Respond(w, r, http.StatusOK, respond{
			Data: found.UserActions,
			Meta: meta{
				TotalPages:      found.TotalPages,
				NumberOfResults: found.NumberOfResults,
			},
		})
	}
}
