package controller

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/internal/app/service"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/api"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/cookie"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/e"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/email"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/flash"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/ip"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/jwt"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/l"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/log"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/recaptcha"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/template"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/util"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/validate"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type userHandler struct {
	once *sync.Once
}

var UserHandler = newUserHandler()

func newUserHandler() *userHandler {
	return &userHandler{
		once: new(sync.Once),
	}
}

func (u *userHandler) RegisterRoutes(
	public *mux.Router,
	private *mux.Router,
	adminPublic *mux.Router,
	adminPrivate *mux.Router,
) {
	u.once.Do(func() {
		public.Path("/lost-password").HandlerFunc(u.lostPasswordPage()).Methods("GET")
		public.Path("/lost-password").HandlerFunc(u.lostPassword()).Methods("POST")
		public.Path("/password-resets/{token}").HandlerFunc(u.passwordResetPage()).Methods("GET")
		public.Path("/password-resets/{token}").HandlerFunc(u.passwordReset()).Methods("POST")

		public.Path("/api/v1/signup").HandlerFunc(u.signup()).Methods("POST")
		public.Path("/api/v1/login").HandlerFunc(u.login()).Methods("POST")
		private.Path("/api/v1/logout").HandlerFunc(u.logout()).Methods("GET")
		private.Path("/api/v1/users/removeFromFavoriteBusinesses").HandlerFunc(u.removeFromFavoriteBusinesses()).Methods("POST")
		private.Path("/api/v1/users/toggleShowRecentMatchedTags").HandlerFunc(u.toggleShowRecentMatchedTags()).Methods("POST")
		private.Path("/api/v1/users/addToFavoriteBusinesses").HandlerFunc(u.addToFavoriteBusinesses()).Methods("POST")
	})
}

func (u *userHandler) FindByID(id string) (*types.User, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, e.Wrap(err, "controller.User.FindByID failed")
	}
	user, err := service.User.FindByID(objID)
	if err != nil {
		return nil, e.Wrap(err, "controller.User.FindByID failed")
	}
	return user, nil
}

func (u *userHandler) FindByBusinessID(id string) (*types.User, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, e.Wrap(err, "controller.User.FindByBusinessID failed")
	}
	user, err := service.User.FindByBusinessID(objID)
	if err != nil {
		return nil, e.Wrap(err, "controller.User.FindByBusinessID failed")
	}
	return user, nil
}

func (u *userHandler) signup() func(http.ResponseWriter, *http.Request) {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type data struct {
		Token string `json:"token"`
	}
	type response struct {
		Data data `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			l.Logger.Error("[ERROR] UserHandler.signup failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		errs := validate.SignUp(req.Email, req.Password)
		if service.User.UserEmailExists(req.Email) {
			errs = append(errs, "Email address is already registered.")
		}
		if len(errs) > 0 {
			l.Logger.Info("[ERROR] UserHandler.signup failed", zap.Strings("input invalid", errs))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		userID, err := service.User.Create(req.Email, req.Password)
		if err != nil {
			l.Logger.Error("[ERROR] UserHandler.signup failed", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		token, err := jwt.GenerateToken(userID, false)
		if err != nil {
			l.Logger.Error("[ERROR] UserHandler.signup failed", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}
		http.SetCookie(w, cookie.CreateCookie(token))

		api.Respond(w, r, http.StatusOK)
	}
}

func (u *userHandler) login() func(http.ResponseWriter, *http.Request) {
	t := template.NewView("account")
	type formData struct {
		Email            string
		Password         string
		RecaptchaSitekey string
		RedirectURL      string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		f := formData{
			Email:            r.FormValue("email"),
			Password:         r.FormValue("password"),
			RecaptchaSitekey: viper.GetString("recaptcha.site_key"),
			RedirectURL:      r.URL.Query().Get("redirect_login"),
		}

		user, err := service.User.Login(f.Email, f.Password)
		if err != nil {
			l.Logger.Info("LoginHandler failed", zap.Error(err))

			// Logic to update user login attempts.
			passwordInvalid := e.IsPasswordInvalid(err)
			if passwordInvalid {
				err := service.User.UpdateLoginAttempts(f.Email)
				if err != nil {
					l.Logger.Error("UpdateLoginAttempts failed", zap.Error(err))
				}
			}

			t.Error(w, r, f, err)

			go func() {
				user, err := service.User.FindByEmail(f.Email)
				if err != nil {
					if !e.IsUserNotFound(err) {
						l.Logger.Error("log.User.LoginFailure failed", zap.Error(err))
					}
					return
				}
				err = service.UserAction.Log(log.User.LoginFailure(user, ip.FromRequest(r)))
				if err != nil {
					l.Logger.Error("log.User.LoginFailure failed", zap.Error(err))
				}
			}()
			return
		}

		token, err := jwt.GenerateToken(user.ID.Hex(), false)
		http.SetCookie(w, cookie.CreateCookie(token))

		// CurrentLoginDate and CurrentLoginIP are the previous informations.
		flash.Info(w, "You last logged in on "+util.FormatTime(user.CurrentLoginDate)+" from "+user.CurrentLoginIP)

		go func() {
			err := service.User.UpdateLoginInfo(user.ID, ip.FromRequest(r))
			if err != nil {
				l.Logger.Error("UpdateLoginInfo failed", zap.Error(err))
			}
		}()
		go func() {
			err := service.UserAction.Log(log.User.LoginSuccess(user, ip.FromRequest(r)))
			if err != nil {
				l.Logger.Error("log.User.LoginSuccess failed", zap.Error(err))
			}
		}()

		http.Redirect(w, r, r.URL.Query().Get("redirect_login"), http.StatusFound)
	}
}

// LogoutHandler logs out the user.
func (u *userHandler) logout() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, cookie.ResetCookie())
		api.Respond(w, r, http.StatusOK)
	}
}

// LostPasswordPage renders the lost password page.
func (u *userHandler) lostPasswordPage() func(http.ResponseWriter, *http.Request) {
	t := template.NewView("lost-password")
	type formData struct {
		Email            string
		Success          bool
		RecaptchaSitekey string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		t.Render(w, r, formData{RecaptchaSitekey: viper.GetString("recaptcha.site_key")}, nil)
	}
}

// LostPassword sends the reset password email.
func (u *userHandler) lostPassword() func(http.ResponseWriter, *http.Request) {
	t := template.NewView("lost-password")
	type formData struct {
		Email            string
		Success          bool
		RecaptchaSitekey string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		f := formData{
			Email:            r.FormValue("email"),
			RecaptchaSitekey: viper.GetString("recaptcha.site_key"),
		}

		if viper.GetString("env") == "production" {
			isValid := recaptcha.Verify(*r)
			if !isValid {
				l.Logger.Info("LostPassword failed", zap.Strings("errs", recaptcha.Error()))
				t.Render(w, r, f, recaptcha.Error())
				return
			}
		}

		user, err := service.User.FindByEmail(f.Email)
		if err != nil {
			l.Logger.Info("LostPassword failed", zap.Error(err))
			t.Error(w, r, f, err)
			return
		}

		receiver := user.FirstName + " " + user.LastName
		lostPassword, err := service.Lostpassword.FindByEmail(f.Email)
		if err == nil && !service.Lostpassword.TokenInvalid(lostPassword) {
			email.SendResetEmail(receiver, f.Email, lostPassword.Token)
			f.Success = true
			t.Render(w, r, f, nil)
			return
		}

		uid, err := uuid.NewV4()
		if err != nil {
			l.Logger.Error("LostPassword failed", zap.Error(err))
			t.Error(w, r, f, err)
			return
		}

		lostPassword = &types.LostPassword{
			Email: user.Email,
			Token: uid.String(),
		}
		err = service.Lostpassword.Create(lostPassword)
		if err != nil {
			l.Logger.Error("LostPassword failed", zap.Error(err))
			t.Error(w, r, f, err)
			return
		}

		email.SendResetEmail(receiver, f.Email, uid.String())

		go func() {
			err := service.UserAction.Log(log.User.LostPassword(user))
			if err != nil {
				l.Logger.Error("log.User.LostPassword failed", zap.Error(err))
			}
		}()

		f.Success = true
		t.Render(w, r, f, nil)
	}
}

// PasswordResetPage renders the lost password page.
func (u *userHandler) passwordResetPage() func(http.ResponseWriter, *http.Request) {
	t := template.NewView("password-resets")
	type formData struct {
		Token string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		token := vars["token"]
		lostPassword, err := service.Lostpassword.FindByToken(token)
		if err != nil {
			l.Logger.Error("PasswordResetPage failed", zap.Error(err))
			http.Redirect(w, r, "/lost-password", http.StatusFound)
			return
		}
		if service.Lostpassword.TokenInvalid(lostPassword) {
			l.Logger.Info("PasswordResetPage failed: token expired \n")
			http.Redirect(w, r, "/lost-password", http.StatusFound)
			return
		}

		t.Render(w, r, formData{Token: token}, nil)
	}
}

// PasswordReset resets the password for a user.
func (u *userHandler) passwordReset() func(http.ResponseWriter, *http.Request) {
	t := template.NewView("password-resets")
	type formData struct {
		Token           string
		Password        string
		ConfirmPassword string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		vars := mux.Vars(r)
		f := formData{
			Token:           vars["token"],
			Password:        r.FormValue("password"),
			ConfirmPassword: r.FormValue("confirm_password"),
		}

		errorMessages := validate.ValidatePassword(f.Password, f.ConfirmPassword)
		if len(errorMessages) > 0 {
			l.Logger.Error("PasswordReset failed", zap.Strings("input invalid", errorMessages))
			t.Render(w, r, f, errorMessages)
			return
		}

		lost, err := service.Lostpassword.FindByToken(f.Token)
		if err != nil {
			l.Logger.Error("PasswordReset failed", zap.Error(err))
			t.Error(w, r, f, err)
			return
		}

		err = service.User.ResetPassword(lost.Email, f.Password)
		if err != nil {
			l.Logger.Error("PasswordReset failed", zap.Error(err))
			t.Error(w, r, f, err)
			return
		}

		go func() {
			err := service.Lostpassword.SetTokenUsed(f.Token)
			if err != nil {
				l.Logger.Error("SetTokenUsed failed", zap.Error(err))
			}
		}()

		go func() {
			user, err := service.User.FindByEmail(lost.Email)
			if err != nil {
				l.Logger.Error("BuildChangePasswordAction failed", zap.Error(err))
				return
			}
			service.UserAction.Log(log.User.ChangePassword(user))
			if err != nil {
				l.Logger.Error("log.User.ChangePassword failed", zap.Error(err))
			}
		}()

		flash.Success(w, "Your password has been reset successfully!")
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func (u *userHandler) toggleShowRecentMatchedTags() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		objID, err := primitive.ObjectIDFromHex(r.Header.Get("userID"))
		if err != nil {
			l.Logger.Error("ToggleShowRecentMatchedTags failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = service.User.ToggleShowRecentMatchedTags(objID)
		if err != nil {
			l.Logger.Error("ToggleShowRecentMatchedTags failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (u *userHandler) addToFavoriteBusinesses() func(http.ResponseWriter, *http.Request) {
	type request struct {
		ID string `json:"id"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil || req.ID == "" {
			if err != nil {
				l.Logger.Error("AppServer AddToFavoriteBusinesses failed", zap.Error(err))
			} else {
				l.Logger.Error("AppServer AddToFavoriteBusinesses failed", zap.String("error", "request business id is empty"))
			}
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}
		bID, err := primitive.ObjectIDFromHex(req.ID)
		if err != nil {
			l.Logger.Error("AppServer AddToFavoriteBusinesses failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}

		uID, err := primitive.ObjectIDFromHex(r.Header.Get("userID"))
		if err != nil {
			l.Logger.Error("AppServer AddToFavoriteBusinesses failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}

		err = service.User.AddToFavoriteBusinesses(uID, bID)
		if err != nil {
			l.Logger.Error("AppServer AddToFavoriteBusinesses failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (u *userHandler) removeFromFavoriteBusinesses() func(http.ResponseWriter, *http.Request) {
	type request struct {
		ID string `json:"id"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil || req.ID == "" {
			if err != nil {
				l.Logger.Error("AppServer RemoveFromFavoriteBusinesses failed", zap.Error(err))
			} else {
				l.Logger.Error("AppServer RemoveFromFavoriteBusinesses failed", zap.String("error", "request business id is empty"))
			}
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}
		bID, err := primitive.ObjectIDFromHex(req.ID)
		if err != nil {
			l.Logger.Error("AppServer RemoveFromFavoriteBusinesses failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}

		uID, err := primitive.ObjectIDFromHex(r.Header.Get("userID"))
		if err != nil {
			l.Logger.Error("AppServer RemoveFromFavoriteBusinesses failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}

		err = service.User.RemoveFromFavoriteBusinesses(uID, bID)
		if err != nil {
			l.Logger.Error("AppServer RemoveFromFavoriteBusinesses failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
