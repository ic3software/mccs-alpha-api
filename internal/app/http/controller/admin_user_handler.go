package controller

import (
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/internal/app/api"
	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/cookie"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/jwt"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/l"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/template"
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

func (a *adminUserHandler) RegisterRoutes(
	public *mux.Router,
	private *mux.Router,
	adminPublic *mux.Router,
	adminPrivate *mux.Router,
) {
	a.once.Do(func() {
		adminPublic.Path("/login").HandlerFunc(a.login()).Methods("POST")

		adminPrivate.Path("/logout").HandlerFunc(a.logoutHandler()).Methods("GET")
		adminPrivate.Path("/users/{id}").HandlerFunc(a.userPage()).Methods("GET")
	})
}

func (handler *adminUserHandler) login() func(http.ResponseWriter, *http.Request) {
	type data struct {
		Token string `json:"token"`
	}
	type respond struct {
		Data data `json:"data"`
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
			l.Logger.Info("[INFO] UserHandler.login failed", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		token, err := jwt.GenerateToken(user.ID.Hex(), true)

		api.Respond(w, r, http.StatusOK, respond{Data: data{Token: token}})
	}
}

// TO BE REMOVED

func (a *adminUserHandler) logoutHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, cookie.ResetCookie())
		http.Redirect(w, r, "/admin/login", http.StatusFound)
	}
}

func (a *adminUserHandler) userPage() func(http.ResponseWriter, *http.Request) {
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
