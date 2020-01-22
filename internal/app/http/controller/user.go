package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/api"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/cookie"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/e"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/email"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/flash"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/jwt"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/l"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/log"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/recaptcha"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/template"
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
		public.Path("/lost-password").HandlerFunc(u.lostPassword()).Methods("POST")
		public.Path("/password-resets/{token}").HandlerFunc(u.passwordResetPage()).Methods("GET")
		public.Path("/password-resets/{token}").HandlerFunc(u.passwordReset()).Methods("POST")

		public.Path("/api/v1/login").HandlerFunc(u.login()).Methods("POST")
		public.Path("/api/v1/signup").HandlerFunc(u.signup()).Methods("POST")
		private.Path("/api/v1/logout").HandlerFunc(u.logout()).Methods("POST")

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
	user, err := logic.User.FindByID(objID)
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
	user, err := logic.User.FindByBusinessID(objID)
	if err != nil {
		return nil, e.Wrap(err, "controller.User.FindByBusinessID failed")
	}
	return user, nil
}

func (u *userHandler) login() func(http.ResponseWriter, *http.Request) {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			l.Logger.Error("[ERROR] UserHandler.login failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		user, err := logic.User.Login(req.Email, req.Password)
		if err != nil {
			l.Logger.Info("[ERROR] UserHandler.login failed", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		token, err := jwt.GenerateToken(user.ID.Hex(), false)
		http.SetCookie(w, cookie.CreateCookie(token))

		api.Respond(w, r, http.StatusOK)
	}
}

func (u *userHandler) signup() func(http.ResponseWriter, *http.Request) {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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
		if logic.User.UserEmailExists(req.Email) {
			errs = append(errs, errors.New("Email address is already registered."))
		}
		if len(errs) > 0 {
			l.Logger.Info("[ERROR] UserHandler.signup failed", zap.Errors("input invalid", errs))
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		userID, err := logic.User.Create(req.Email, req.Password)
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

// LogoutHandler logs out the user.
func (u *userHandler) logout() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, cookie.ResetCookie())
		api.Respond(w, r, http.StatusOK)
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

		user, err := logic.User.FindByEmail(f.Email)
		if err != nil {
			l.Logger.Info("LostPassword failed", zap.Error(err))
			t.Error(w, r, f, err)
			return
		}

		receiver := user.FirstName + " " + user.LastName
		lostPassword, err := logic.Lostpassword.FindByEmail(f.Email)
		if err == nil && !logic.Lostpassword.TokenInvalid(lostPassword) {
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
		err = logic.Lostpassword.Create(lostPassword)
		if err != nil {
			l.Logger.Error("LostPassword failed", zap.Error(err))
			t.Error(w, r, f, err)
			return
		}

		email.SendResetEmail(receiver, f.Email, uid.String())

		go func() {
			err := logic.UserAction.Log(log.User.LostPassword(user))
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
		lostPassword, err := logic.Lostpassword.FindByToken(token)
		if err != nil {
			l.Logger.Error("PasswordResetPage failed", zap.Error(err))
			http.Redirect(w, r, "/lost-password", http.StatusFound)
			return
		}
		if logic.Lostpassword.TokenInvalid(lostPassword) {
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

		lost, err := logic.Lostpassword.FindByToken(f.Token)
		if err != nil {
			l.Logger.Error("PasswordReset failed", zap.Error(err))
			t.Error(w, r, f, err)
			return
		}

		err = logic.User.ResetPassword(lost.Email, f.Password)
		if err != nil {
			l.Logger.Error("PasswordReset failed", zap.Error(err))
			t.Error(w, r, f, err)
			return
		}

		go func() {
			err := logic.Lostpassword.SetTokenUsed(f.Token)
			if err != nil {
				l.Logger.Error("SetTokenUsed failed", zap.Error(err))
			}
		}()

		go func() {
			user, err := logic.User.FindByEmail(lost.Email)
			if err != nil {
				l.Logger.Error("BuildChangePasswordAction failed", zap.Error(err))
				return
			}
			logic.UserAction.Log(log.User.ChangePassword(user))
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
		err = logic.User.ToggleShowRecentMatchedTags(objID)
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

		err = logic.User.AddToFavoriteBusinesses(uID, bID)
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

		err = logic.User.RemoveFromFavoriteBusinesses(uID, bID)
		if err != nil {
			l.Logger.Error("AppServer RemoveFromFavoriteBusinesses failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
