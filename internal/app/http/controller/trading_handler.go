package controller

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/global/constant"
	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/email"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/helper"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/l"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/template"
	"go.uber.org/zap"
)

type tradingHandler struct {
	once *sync.Once
}

var TradingHandler = newTradingHandler()

func newTradingHandler() *tradingHandler {
	return &tradingHandler{
		once: new(sync.Once),
	}
}

func (th *tradingHandler) RegisterRoutes(
	public *mux.Router,
	private *mux.Router,
	adminPublic *mux.Router,
	adminPrivate *mux.Router,
) {
	th.once.Do(func() {
		private.Path("/member-signup").HandlerFunc(th.signup()).Methods("POST")

		private.Path("/api/is-trading-member").HandlerFunc(th.isMember()).Methods("GET")
	})
}

func (th *tradingHandler) signup() func(http.ResponseWriter, *http.Request) {
	t := template.NewView("member-signup")
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		// Validate user inputs.
		data := helper.Trading.GetRegisterData(r)
		errorMessages := data.Validate()
		if len(errorMessages) > 0 {
			l.Logger.Info("TradingHandler.Signup failed", zap.Strings("input invalid", errorMessages))
			t.Render(w, r, data, errorMessages)
			return
		}

		// Update entity collection.
		entity, err := EntityHandler.FindByUserID(r.Header.Get("userID"))
		if err != nil {
			l.Logger.Info("TradingHandler.Signup failed", zap.Error(err))
			t.Error(w, r, data, err)
			return
		}
		err = logic.Trading.UpdateEntity(entity.ID, data)
		if err != nil {
			l.Logger.Info("TradingHandler.Signup failed", zap.Error(err))
			t.Error(w, r, data, err)
			return
		}

		// Update user collection.
		user, err := UserHandler.FindByID(r.Header.Get("userID"))
		if err != nil {
			l.Logger.Info("TradingHandler.Signup failed", zap.Error(err))
			t.Error(w, r, data, err)
			return
		}
		err = logic.Trading.UpdateUser(user.ID, data)
		if err != nil {
			l.Logger.Info("TradingHandler.Signup failed", zap.Error(err))
			t.Error(w, r, data, err)
			return
		}

		// Send thank you email to the User's email address.
		go func() {
			err := email.SendThankYouEmail(data.FirstName, data.LastName, user.Email)
			if err != nil {
				l.Logger.Error("email.SendThankYouEmail failed", zap.Error(err))
			}
		}()
		// Send the to the OCN Admin email address.
		go func() {
			err := email.SendNewMemberSignupEmail(data.EntityName, user.Email)
			if err != nil {
				l.Logger.Error("email.SendNewMemberSignupEmail failed", zap.Error(err))
			}
		}()

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func (th *tradingHandler) isMember() func(http.ResponseWriter, *http.Request) {
	type response struct {
		IsMember bool
	}
	return func(w http.ResponseWriter, r *http.Request) {
		entity, err := EntityHandler.FindByUserID(r.Header.Get("userID"))
		if err != nil {
			l.Logger.Error("TradingHandler.IsMember failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		res := response{IsMember: entity.Status == constant.Trading.Accepted}
		js, err := json.Marshal(res)
		if err != nil {
			l.Logger.Error("TradingHandler.IsMember failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}
