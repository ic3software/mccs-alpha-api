package controller

import (
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/cookie"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/e"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/ip"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/jwt"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/l"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/log"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/template"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
		adminPrivate.Path("").HandlerFunc(a.dashboardPage()).Methods("GET")
		adminPublic.Path("/login").HandlerFunc(a.loginHandler()).Methods("POST")
		adminPrivate.Path("/logout").HandlerFunc(a.logoutHandler()).Methods("GET")
		adminPrivate.Path("/users/{id}").HandlerFunc(a.userPage()).Methods("GET")
	})
}

func (a *adminUserHandler) dashboardPage() func(http.ResponseWriter, *http.Request) {
	t := template.NewView("admin/dashboard")
	return func(w http.ResponseWriter, r *http.Request) {
		objID, err := primitive.ObjectIDFromHex(r.Header.Get("userID"))
		if err != nil {
			l.Logger.Error("AdminDashboardPage failed", zap.Error(err))
			t.Error(w, r, nil, err)
			return
		}
		adminUser, err := logic.AdminUser.FindByID(objID)
		if err != nil {
			l.Logger.Error("AdminDashboardPage failed", zap.Error(err))
			t.Error(w, r, nil, err)
			return
		}
		t.Render(w, r, adminUser, nil)
	}
}

func (a *adminUserHandler) loginHandler() func(http.ResponseWriter, *http.Request) {
	t := template.NewView("/admin/login")
	type formData struct {
		Email            string
		Password         string
		RecaptchaSitekey string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		f := formData{
			Email:            r.FormValue("email"),
			Password:         r.FormValue("password"),
			RecaptchaSitekey: viper.GetString("recaptcha.site_key"),
		}

		user, err := logic.AdminUser.Login(f.Email, f.Password)
		if err != nil {
			l.Logger.Info("AdminLoginHandler failed", zap.Error(err))
			t.Error(w, r, f, err)
			go func() {
				user, err := logic.AdminUser.FindByEmail(f.Email)
				if err != nil {
					if !e.IsUserNotFound(err) {
						l.Logger.Error("BuildLoginFailureAction failed", zap.Error(err))
					}
					return
				}
				err = logic.UserAction.Log(log.Admin.LoginFailure(user, ip.FromRequest(r)))
				if err != nil {
					l.Logger.Error("BuildLoginFailureAction failed", zap.Error(err))
				}
			}()
			return
		}

		token, err := jwt.GenerateToken(user.ID.Hex(), true)
		http.SetCookie(w, cookie.CreateCookie(token))

		go func() {
			err := logic.AdminUser.UpdateLoginInfo(user.ID, ip.FromRequest(r))
			if err != nil {
				l.Logger.Error("AdminLoginHandler failed", zap.Error(err))
			}
		}()
		go func() {
			err := logic.UserAction.Log(log.Admin.LoginSuccess(user, ip.FromRequest(r)))
			if err != nil {
				l.Logger.Error("log.Admin.LoginSuccess failed", zap.Error(err))
			}
		}()

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

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
