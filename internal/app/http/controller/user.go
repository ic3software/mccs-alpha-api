package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/internal/app/api"
	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/cookie"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/e"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/email"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/jwt"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/l"
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

func (handler *userHandler) RegisterRoutes(
	public *mux.Router,
	private *mux.Router,
	adminPublic *mux.Router,
	adminPrivate *mux.Router,
) {
	handler.once.Do(func() {
		public.Path("/login").HandlerFunc(handler.login()).Methods("POST")
		public.Path("/signup").HandlerFunc(handler.signup()).Methods("POST")
		private.Path("/logout").HandlerFunc(handler.logout()).Methods("POST")

		public.Path("/password-reset").HandlerFunc(handler.requestPasswordReset()).Methods("POST")
		public.Path("/password-reset/{token}").HandlerFunc(handler.passwordReset()).Methods("POST")
		private.Path("/password-change").HandlerFunc(handler.passwordChange()).Methods("POST")

		private.Path("/user").HandlerFunc(handler.userProfile()).Methods("GET")
		private.Path("/user").HandlerFunc(handler.updateUser()).Methods("PATCH")
		private.Path("/user/entities").HandlerFunc(handler.listUserEntities()).Methods("GET")
		private.Path("/user/entities/{entityID}").HandlerFunc(handler.updateUserEntity()).Methods("PATCH")

		private.Path("/users/toggleShowRecentMatchedTags").HandlerFunc(handler.toggleShowRecentMatchedTags()).Methods("POST")
	})
}

func (handler *userHandler) FindByID(id string) (*types.User, error) {
	objID, _ := primitive.ObjectIDFromHex(id)
	user, err := logic.User.FindByID(objID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (handler *userHandler) FindByEntityID(id string) (*types.User, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, e.Wrap(err, "controller.User.FindByEntityID failed")
	}
	user, err := logic.User.FindByEntityID(objID)
	if err != nil {
		return nil, e.Wrap(err, "controller.User.FindByEntityID failed")
	}
	return user, nil
}

func (handler *userHandler) IsEntityBelongsToUser(entityID, userID string) bool {
	uID, _ := primitive.ObjectIDFromHex(userID)
	user, err := logic.User.FindByID(uID)
	if err != nil {
		return false
	}
	for _, entity := range user.Entities {
		if entity.Hex() == entityID {
			return true
		}
	}
	return false
}

func (u *userHandler) updateLoginAttempts(email string) {
	err := logic.User.UpdateLoginAttempts(email)
	if err != nil {
		l.Logger.Error("[Error] UserHandler.updateLoginAttempts failed:", zap.Error(err))
	}
}

func (handler *userHandler) login() func(http.ResponseWriter, *http.Request) {
	type data struct {
		Token string `json:"token"`
	}
	type respond struct {
		Data data `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := api.NewLoginReqBody(r)
		if err != nil {
			l.Logger.Info("[INFO] UserHandler.login failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		errs := req.Validate()
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		user, err := logic.User.Login(req.Email, req.Password)
		if err != nil {
			l.Logger.Info("[INFO] UserHandler.login failed", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			go handler.updateLoginAttempts(req.Email)
			return
		}

		token, err := jwt.GenerateToken(user.ID.Hex(), false)

		api.Respond(w, r, http.StatusOK, respond{Data: data{Token: token}})
	}
}

func (handler *userHandler) signup() func(http.ResponseWriter, *http.Request) {
	type data struct {
		UserID   string `json:"userID"`
		EntityID string `json:"entityID"`
		Token    string `json:"token"`
	}
	type respond struct {
		Data data `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := api.NewSignupReqBody(r)
		if err != nil {
			l.Logger.Info("[INFO] UserHandler.signup failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		errs := req.Validate()
		if logic.User.UserEmailExists(req.Email) {
			errs = append(errs, errors.New("Email address is already registered."))
		}
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		entityID, err := logic.Entity.Create(&types.Entity{
			EntityName:         req.EntityName,
			Email:              req.Email,
			IncType:            req.IncType,
			CompanyNumber:      req.CompanyNumber,
			EntityPhone:        req.EntityPhone,
			Website:            req.Website,
			Turnover:           req.Turnover,
			Description:        req.Description,
			LocationAddress:    req.LocationAddress,
			LocationCity:       req.LocationCity,
			LocationRegion:     req.LocationRegion,
			LocationPostalCode: req.LocationPostalCode,
			LocationCountry:    req.LocationCountry,
			Offers:             types.ToTagFields(req.Offers),
			Wants:              types.ToTagFields(req.Wants),
		})
		if err != nil {
			l.Logger.Error("[ERROR] UserHandler.signup failed", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}
		userID, err := logic.User.Create(&types.User{
			Email:                 req.Email,
			Password:              req.Password,
			FirstName:             req.FirstName,
			LastName:              req.LastName,
			Telephone:             req.UserPhone,
			ShowRecentMatchedTags: req.ShowRecentMatchedTags,
			DailyNotification:     req.DailyNotification,
		})
		if err != nil {
			l.Logger.Error("[ERROR] UserHandler.signup failed", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}
		err = logic.Entity.AssociateUser(entityID, userID)
		if err != nil {
			l.Logger.Error("[ERROR] UserHandler.signup failed", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}
		err = logic.User.AssociateEntity(userID, entityID)
		if err != nil {
			l.Logger.Error("[ERROR] UserHandler.signup failed", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		token, err := jwt.GenerateToken(userID.Hex(), false)
		if err != nil {
			l.Logger.Error("[ERROR] UserHandler.signup failed", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		api.Respond(w, r, http.StatusOK, respond{Data: data{
			UserID:   userID.Hex(),
			EntityID: entityID.Hex(),
			Token:    token,
		}})
	}
}

func (handler *userHandler) logout() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, cookie.ResetCookie())
		api.Respond(w, r, http.StatusOK)
	}
}

func (handler *userHandler) requestPasswordReset() func(http.ResponseWriter, *http.Request) {
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

func (handler *userHandler) passwordReset() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		var req types.ResetPasswordReqBody
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			l.Logger.Info("[INFO] UserHandler.passwordReset failed:", zap.Error(err))
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

func (handler *userHandler) passwordChange() func(http.ResponseWriter, *http.Request) {
	type request struct {
		Password string `json:"password"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PasswordChange
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			l.Logger.Info("[INFO] UserHandler.passwordChange failed:", zap.Error(err))
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

func (handler *userHandler) userProfile() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.UserRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := handler.FindByID(r.Header.Get("userID"))
		if err != nil {
			l.Logger.Error("[ERROR] UserHandler.userProfile failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}
		api.Respond(w, r, http.StatusOK, respond{Data: api.NewUserRespond(user)})
	}
}

func (handler *userHandler) updateUser() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.UserRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateUserReqBody
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			l.Logger.Info("[INFO] UserHandler.updateUser failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		errs := req.Validate()
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		userID, _ := primitive.ObjectIDFromHex(r.Header.Get("userID"))
		user, err := logic.User.FindOneAndUpdate(&types.User{
			ID:                    userID,
			FirstName:             req.FirstName,
			LastName:              req.LastName,
			Telephone:             req.UserPhone,
			DailyNotification:     req.DailyEmailMatchNotification,
			ShowRecentMatchedTags: req.ShowTagsMatchedSinceLastLogin,
		})
		if err != nil {
			l.Logger.Info("[INFO] UserHandler.updateUser failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		api.Respond(w, r, http.StatusOK, respond{Data: api.NewUserRespond(user)})
	}
}

func (handler *userHandler) listUserEntities() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data []*types.EntityRespond `json:"data"`
	}
	toData := func(entities []*types.Entity) []*types.EntityRespond {
		result := []*types.EntityRespond{}
		for _, entity := range entities {
			result = append(result, api.NewEntityRespondWithEmail(entity))
		}
		return result
	}
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := primitive.ObjectIDFromHex(r.Header.Get("userID"))
		entities, err := logic.User.FindEntities(userID)
		if err != nil {
			l.Logger.Error("[Error] UserHandler.listUserEntities failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}
		api.Respond(w, r, http.StatusOK, respond{Data: toData(entities)})
	}
}

func (handler *userHandler) updateUserEntity() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.EntityRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := api.NewUpdateUserEntityReqBody(r)
		if err != nil {
			l.Logger.Info("[INFO] UserHandler.updateUserEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		errs := req.Validate()
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		vars := mux.Vars(r)
		if !handler.IsEntityBelongsToUser(vars["entityID"], r.Header.Get("userID")) {
			api.Respond(w, r, http.StatusForbidden, api.ErrPermissionDenied)
			return
		}

		entityID, _ := primitive.ObjectIDFromHex(vars["entityID"])
		oldEntity, err := logic.Entity.FindByID(entityID)
		if err != nil {
			l.Logger.Error("[Error] UserHandler.updateUserEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		entity, err := logic.Entity.FindOneAndUpdate(&types.Entity{
			ID:                 entityID,
			EntityName:         req.EntityName,
			Email:              req.Email,
			EntityPhone:        req.EntityPhone,
			IncType:            req.IncType,
			CompanyNumber:      req.CompanyNumber,
			Website:            req.Website,
			Turnover:           req.Turnover,
			Description:        req.Description,
			LocationAddress:    req.LocationAddress,
			LocationCity:       req.LocationCity,
			LocationRegion:     req.LocationRegion,
			LocationPostalCode: req.LocationPostalCode,
			LocationCountry:    req.LocationCountry,
		})
		if err != nil {
			l.Logger.Info("[INFO] UserHandler.updateUserEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		go EntityHandler.UpdateOffersAndWants(oldEntity, req.Offers, req.Wants)

		if len(req.Offers) != 0 {
			entity.Offers = types.ToTagFields(req.Offers)
		}
		if len(req.Wants) != 0 {
			entity.Wants = types.ToTagFields(req.Wants)
		}
		api.Respond(w, r, http.StatusOK, respond{Data: api.NewEntityRespondWithEmail(entity)})
	}
}

func (handler *userHandler) toggleShowRecentMatchedTags() func(http.ResponseWriter, *http.Request) {
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
