package controller

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/helper"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/l"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/template"
	"go.uber.org/zap"
)

type dashBoardHandler struct {
	once *sync.Once
}

var DashBoardHandler = newDashBoardHandler()

func newDashBoardHandler() *dashBoardHandler {
	return &dashBoardHandler{
		once: new(sync.Once),
	}
}

func (d *dashBoardHandler) RegisterRoutes(
	public *mux.Router,
	private *mux.Router,
	adminPublic *mux.Router,
	adminPrivate *mux.Router,
) {
	d.once.Do(func() {
		private.Path("/").HandlerFunc(d.dashboardPage()).Methods("GET")
	})
}

func (d *dashBoardHandler) dashboardPage() func(http.ResponseWriter, *http.Request) {
	t := template.NewView("dashboard")
	type response struct {
		User          *types.User
		Entity        *types.Entity
		MatchedOffers map[string][]string
		MatchedWants  map[string][]string
		Balance       float64
	}
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := UserHandler.FindByID(r.Header.Get("userID"))

		if err != nil {
			l.Logger.Error("DashboardPage failed", zap.Error(err))
			t.Error(w, r, nil, err)
			return
		}

		entity, err := logic.Entity.FindByID(user.Entities[0])
		if err != nil {
			l.Logger.Error("DashboardPage failed", zap.Error(err))
			t.Error(w, r, nil, err)
			return
		}

		lastLoginDate := time.Time{}
		if user.ShowRecentMatchedTags {
			lastLoginDate = user.LastLoginDate
		}

		matchedOffers, err := logic.Tag.MatchOffers(helper.GetTagNames(entity.Offers), lastLoginDate)
		if err != nil {
			l.Logger.Error("DashboardPage failed", zap.Error(err))
			t.Error(w, r, nil, err)
			return
		}
		matchedWants, err := logic.Tag.MatchWants(helper.GetTagNames(entity.Wants), lastLoginDate)
		if err != nil {
			l.Logger.Error("DashboardPage failed", zap.Error(err))
			t.Error(w, r, nil, err)
			return
		}

		res := response{
			User:          user,
			Entity:        entity,
			MatchedOffers: matchedOffers,
			MatchedWants:  matchedWants,
		}

		// Get the account balance.
		account, err := logic.Account.FindByEntityID(user.Entities[0].Hex())
		if err != nil {
			l.Logger.Error("DashboardPage failed", zap.Error(err))
			t.Error(w, r, nil, err)
			return
		}
		res.Balance = account.Balance

		t.Render(w, r, res, nil)
	}
}
