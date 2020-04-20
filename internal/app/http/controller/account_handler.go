package controller

import (
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/global/constant"
	"github.com/ic3network/mccs-alpha-api/internal/app/api"
	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/l"
	"go.uber.org/zap"
)

type accountHandler struct {
	once *sync.Once
}

var AccountHandler = newAccountHandler()

func newAccountHandler() *accountHandler {
	return &accountHandler{
		once: new(sync.Once),
	}
}

func (handler *accountHandler) RegisterRoutes(
	public *mux.Router,
	private *mux.Router,
	adminPublic *mux.Router,
	adminPrivate *mux.Router,
) {
	handler.once.Do(func() {
		private.Path("/balance").HandlerFunc(handler.getBalance()).Methods("GET")
	})
}

func (handler *accountHandler) getBalance() func(http.ResponseWriter, *http.Request) {
	type data struct {
		Unit    string  `json:"unit"`
		Balance float64 `json:"balance"`
	}
	type respond struct {
		Data data `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		query, errs := types.NewBalanceQuery(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		if !UserHandler.IsEntityBelongsToUser(query.QueryingEntityID, r.Header.Get("userID")) {
			api.Respond(w, r, http.StatusForbidden, api.ErrPermissionDenied)
			return
		}

		account, err := logic.Account.FindByEntityID(query.QueryingEntityID)
		if err != nil {
			l.Logger.Error("[Error] Account.getBalance failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		api.Respond(w, r, http.StatusOK, respond{Data: data{
			Unit:    constant.Unit.UK,
			Balance: account.Balance,
		}})
	}
}
