package controller

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/internal/app/api"
	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/cookie"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/jwt"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/l"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/template"
	"github.com/ic3network/mccs-alpha-api/util"
	"go.uber.org/zap"
)

type adminUserHandler struct {
	once *sync.Once
}

var AdminUserHandler = newAdminUserHandler()

func newAdminUserHandler() *adminUserHandler {
	return &adminUserHandler{
		once: new(sync.Once),
	}
}

func (handler *adminUserHandler) RegisterRoutes(
	public *mux.Router,
	private *mux.Router,
	adminPublic *mux.Router,
	adminPrivate *mux.Router,
) {
	handler.once.Do(func() {
		adminPublic.Path("/login").HandlerFunc(handler.login()).Methods("POST")
		adminPrivate.Path("/logout").HandlerFunc(handler.logout()).Methods("POST")

		adminPrivate.Path("/users/{id}").HandlerFunc(handler.userPage()).Methods("GET")
	})
}

func (handler *adminUserHandler) login() func(http.ResponseWriter, *http.Request) {
	type data struct {
		Token         string     `json:"token"`
		LastLoginIP   string     `json:"lastLoginIP,omitempty"`
		LastLoginDate *time.Time `json:"lastLoginDate,omitempty"`
	}
	type respond struct {
		Data data `json:"data"`
	}
	respondData := func(user *types.AdminUser, token string) data {
		d := data{Token: token}
		if user.LastLoginIP != "" {
			d.LastLoginIP = user.LastLoginIP
		}
		if !user.LastLoginDate.IsZero() {
			d.LastLoginDate = &user.LastLoginDate
		}
		return d
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := api.NewLoginReqBody(r)
		if err != nil {
			l.Logger.Info("[INFO] AdminUserHandler.login failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		errs := req.Validate()
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		user, err := logic.AdminUser.Login(req.Email, req.Password)
		if err != nil {
			l.Logger.Error("[Error] AdminUserHandler.login failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		token, err := jwt.GenerateToken(user.ID.Hex(), true)

		api.Respond(w, r, http.StatusOK, respond{Data: respondData(user, token)})

		go func() {
			err := logic.AdminUser.UpdateLoginInfo(user.ID, util.IPAddress(r))
			if err != nil {
				l.Logger.Error("[Error] AdminUser.UpdateLoginInfo failed:", zap.Error(err))
			}
		}()
	}
}

func (handler *adminUserHandler) logout() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, cookie.ResetCookie())
		api.Respond(w, r, http.StatusOK)
	}
}

// TO BE REMOVED

func (handler *adminUserHandler) userPage() func(http.ResponseWriter, *http.Request) {
	t := template.NewView("admin/user")
	type formData struct {
		User *types.User
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userID := vars["id"]
		user, err := UserHandler.FindByID(userID)
		if err != nil {
			l.Logger.Error("UserPage failed", zap.Error(err))
			t.Error(w, r, nil, err)
			return
		}

		f := formData{User: user}

		t.Render(w, r, f, nil)
	}
}
