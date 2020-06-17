package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/internal/app/api"
	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/email"
	"github.com/ic3network/mccs-alpha-api/util"
	"github.com/ic3network/mccs-alpha-api/util/cookie"
	"github.com/ic3network/mccs-alpha-api/util/l"
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

func (handler *adminUserHandler) RegisterRoutes(
	public *mux.Router,
	private *mux.Router,
	adminPublic *mux.Router,
	adminPrivate *mux.Router,
) {
	handler.once.Do(func() {
		adminPublic.Path("/login").HandlerFunc(handler.login()).Methods("POST")
		adminPrivate.Path("/logout").HandlerFunc(handler.logout()).Methods("POST")
		adminPublic.Path("/password-reset").HandlerFunc(handler.requestPasswordReset()).Methods("POST")
		adminPublic.Path("/password-reset/{token}").HandlerFunc(handler.passwordReset()).Methods("POST")
		adminPrivate.Path("/password-change").HandlerFunc(handler.passwordChange()).Methods("POST")
	})
}

// POST /admin/login

func (handler *adminUserHandler) login() func(http.ResponseWriter, *http.Request) {
	type data struct {
		Token         string     `json:"token"`
		LastLoginIP   string     `json:"lastLoginIP,omitempty"`
		LastLoginDate *time.Time `json:"lastLoginDate,omitempty"`
	}
	type respond struct {
		Data data `json:"data"`
	}
	respondData := func(info *types.LoginInfo, token string) data {
		d := data{Token: token}
		if info.LastLoginIP != "" {
			d.LastLoginIP = info.LastLoginIP
		}
		if !info.LastLoginDate.IsZero() {
			d.LastLoginDate = &info.LastLoginDate
		}
		return d
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := types.NewLoginReq(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		user, err := logic.AdminUser.Login(req.Email, req.Password)
		if err != nil {
			l.Logger.Info("[Info] AdminUserHandler.login failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			go logic.AdminUser.IncLoginAttempts(req.Email)
			go logic.UserAction.AdminLoginFail(req.Email, util.IPAddress(r))
			return
		}
		loginInfo, err := logic.AdminUser.UpdateLoginInfo(user.ID, util.IPAddress(r))
		if err != nil {
			l.Logger.Error("[Error] AdminUser.UpdateLoginInfo failed:", zap.Error(err))
		}

		token, err := util.GenerateToken(user.ID.Hex(), true)

		go logic.UserAction.AdminLogin(user, util.IPAddress(r))

		api.Respond(w, r, http.StatusOK, respond{Data: respondData(loginInfo, token)})
	}
}

// POST /admin/logout

func (handler *adminUserHandler) logout() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, cookie.ResetCookie())
		api.Respond(w, r, http.StatusOK)
	}
}

// POST /admin/password-reset

func (handler *adminUserHandler) requestPasswordReset() func(http.ResponseWriter, *http.Request) {
	type request struct {
		Email string `json:"email"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		user, err := logic.AdminUser.FindByEmail(req.Email)
		if err != nil {
			l.Logger.Info("[INFO] AdminUserHandler.requestPasswordReset failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		var token string
		lostPassword, err := logic.Lostpassword.FindByEmail(req.Email)
		if err == nil && logic.Lostpassword.IsTokenValid(lostPassword) {
			token = lostPassword.Token
		} else {
			uid, err := uuid.NewV4()
			if err != nil {
				l.Logger.Error("[ERROR] AdminUserHandler.requestPasswordReset failed:", zap.Error(err))
				api.Respond(w, r, http.StatusInternalServerError, err)
				return
			}
			err = logic.Lostpassword.Create(&types.LostPassword{Email: user.Email, Token: uid.String()})
			if err != nil {
				l.Logger.Error("[ERROR] AdminUserHandler.requestPasswordReset failed:", zap.Error(err))
				api.Respond(w, r, http.StatusInternalServerError, err)
				return
			}
			token = uid.String()
		}

		go email.AdminPasswordReset(&email.AdminPasswordResetEmail{
			Receiver:      user.Name,
			ReceiverEmail: req.Email,
			Token:         token,
		})

		if viper.GetString("env") == "development" {
			type data struct {
				Token string `json:"token"`
			}
			type respond struct {
				Data data `json:"data"`
			}
			api.Respond(w, r, http.StatusOK, respond{Data: data{Token: token}})
			return
		}
		api.Respond(w, r, http.StatusOK)
	}
}

// POST /admin/password-reset/{token}

func (handler *adminUserHandler) passwordReset() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		var req types.ResetPasswordReq
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		errs := req.Validate()
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		lostPassword, err := logic.Lostpassword.FindByToken(vars["token"])
		if err != nil || logic.Lostpassword.IsTokenInvalid(lostPassword) {
			api.Respond(w, r, http.StatusBadRequest, errors.New("Token is invalid."))
			return
		}

		err = logic.AdminUser.ResetPassword(lostPassword.Email, req.Password)
		if err != nil {
			l.Logger.Error("[ERROR] AdminUserHandler.passwordReset failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		go func() {
			err := logic.Lostpassword.SetTokenUsed(vars["token"])
			if err != nil {
				l.Logger.Error("[ERROR] SetTokenUsed failed:", zap.Error(err))
			}
		}()

		api.Respond(w, r, http.StatusOK)
	}
}

// POST /admin/password-change

func (handler *adminUserHandler) passwordChange() func(http.ResponseWriter, *http.Request) {
	type request struct {
		Password string `json:"password"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PasswordChange
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		errs := req.Validate()
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		objID, err := primitive.ObjectIDFromHex(r.Header.Get("userID"))
		if err != nil {
			l.Logger.Error("[ERROR] AdminUserHandler.passwordChange failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}
		user, err := logic.AdminUser.FindByID(objID)
		if err != nil {
			l.Logger.Error("[ERROR] AdminUserHandler.passwordChange failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		err = logic.AdminUser.ResetPassword(user.Email, req.Password)
		if err != nil {
			l.Logger.Error("[ERROR] AdminUserHandler.passwordChange failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		api.Respond(w, r, http.StatusOK)
	}
}
