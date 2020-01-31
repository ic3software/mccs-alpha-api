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
	"github.com/ic3network/mccs-alpha-api/internal/pkg/jwt"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/l"
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
		public.Path("/api/v1/login").HandlerFunc(u.login()).Methods("POST")
		public.Path("/api/v1/signup").HandlerFunc(u.signup()).Methods("POST")
		private.Path("/api/v1/logout").HandlerFunc(u.logout()).Methods("POST")
		public.Path("/api/v1/password-reset").HandlerFunc(u.requestPasswordReset()).Methods("POST")
		public.Path("/api/v1/password-reset/{token}").HandlerFunc(u.passwordReset()).Methods("POST")
		private.Path("/api/v1/password-change").HandlerFunc(u.passwordChange()).Methods("POST")

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
	type data struct {
		Token string `json:"token"`
	}
	type respond struct {
		Data data `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			l.Logger.Info("[INFO] UserHandler.login failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		errs := validate.Login(req.Password)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		user, err := logic.User.Login(req.Email, req.Password)
		if err != nil {
			l.Logger.Info("[INFO] UserHandler.login failed", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		token, err := jwt.GenerateToken(user.ID.Hex(), false)

		api.Respond(w, r, http.StatusOK, respond{Data: data{Token: token}})
	}
}

func (u *userHandler) signup() func(http.ResponseWriter, *http.Request) {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type data struct {
		Token string `json:"token"`
	}
	type respond struct {
		Data data `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			l.Logger.Info("[INFO] UserHandler.signup failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		errs := validate.SignUp(req.Email, req.Password)
		if logic.User.UserEmailExists(req.Email) {
			errs = append(errs, errors.New("User signup failed: Email address is already registered."))
		}
		if len(errs) > 0 {
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

		w.Header().Set("Location", viper.GetString("url")+"/api/v1/users/"+userID)
		api.Respond(w, r, http.StatusCreated, respond{Data: data{Token: token}})
	}
}

func (u *userHandler) logout() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, cookie.ResetCookie())
		api.Respond(w, r, http.StatusOK)
	}
}

func (u *userHandler) requestPasswordReset() func(http.ResponseWriter, *http.Request) {
	type request struct {
		Email string `json:"email"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			l.Logger.Info("[INFO] UserHandler.requestPasswordReset failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		user, err := logic.User.FindByEmail(req.Email)
		if err != nil {
			l.Logger.Info("[INFO] UserHandler.requestPasswordReset failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		receiver := user.FirstName + " " + user.LastName

		lostPassword, err := logic.Lostpassword.FindByEmail(req.Email)
		if err == nil && logic.Lostpassword.IsTokenValid(lostPassword) {
			email.SendResetEmail(receiver, req.Email, lostPassword.Token)
			api.Respond(w, r, http.StatusOK)
			return
		}

		uid, err := uuid.NewV4()
		if err != nil {
			l.Logger.Error("[ERROR] UserHandler.requestPasswordReset failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		err = logic.Lostpassword.Create(&types.LostPassword{Email: user.Email, Token: uid.String()})
		if err != nil {
			l.Logger.Error("[ERROR] UserHandler.requestPasswordReset failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		email.SendResetEmail(receiver, req.Email, uid.String())

		if viper.GetString("env") == "development" {
			type data struct {
				Token string `json:"token"`
			}
			type respond struct {
				Data data `json:"data"`
			}
			api.Respond(w, r, http.StatusOK, respond{Data: data{Token: uid.String()}})
			return
		}
		api.Respond(w, r, http.StatusOK)
	}
}

func (u *userHandler) passwordReset() func(http.ResponseWriter, *http.Request) {
	type request struct {
		Password string `json:"password"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		var req request
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			l.Logger.Info("[INFO] UserHandler.passwordReset failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		errs := validate.ResetPassword(req.Password)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		lostPassword, err := logic.Lostpassword.FindByToken(vars["token"])
		if err != nil || logic.Lostpassword.IsTokenInvalid(lostPassword) {
			api.Respond(w, r, http.StatusBadRequest, errors.New("Password reset failed: Token is invalid."))
			return
		}

		err = logic.User.ResetPassword(lostPassword.Email, req.Password)
		if err != nil {
			l.Logger.Error("[ERROR] UserHandler.passwordReset failed:", zap.Error(err))
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

func (u *userHandler) passwordChange() func(http.ResponseWriter, *http.Request) {
	type request struct {
		Password string `json:"password"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			l.Logger.Info("[INFO] UserHandler.passwordChange failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		errs := validate.ResetPassword(req.Password)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		objID, err := primitive.ObjectIDFromHex(r.Header.Get("userID"))
		if err != nil {
			l.Logger.Error("[ERROR] UserHandler.passwordChange failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}
		user, err := logic.User.FindByID(objID)
		if err != nil {
			l.Logger.Error("[ERROR] UserHandler.passwordChange failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		err = logic.User.ResetPassword(user.Email, req.Password)
		if err != nil {
			l.Logger.Error("[ERROR] UserHandler.passwordChange failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		api.Respond(w, r, http.StatusOK)
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
