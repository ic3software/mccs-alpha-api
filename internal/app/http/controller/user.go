package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

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
		public.Path("/api/v1/login").HandlerFunc(u.login()).Methods("POST")
		public.Path("/api/v1/signup").HandlerFunc(u.signup()).Methods("POST")
		private.Path("/api/v1/logout").HandlerFunc(u.logout()).Methods("POST")
		public.Path("/api/v1/password-reset").HandlerFunc(u.requestPasswordReset()).Methods("POST")
		public.Path("/api/v1/password-reset/{token}").HandlerFunc(u.passwordReset()).Methods("POST")
		private.Path("/api/v1/password-change").HandlerFunc(u.passwordChange()).Methods("POST")
		private.Path("/api/v1/user").HandlerFunc(u.userProfile()).Methods("GET")
		private.Path("/api/v1/user").HandlerFunc(u.updateUser()).Methods("PATCH")
		private.Path("/api/v1/user/entities").HandlerFunc(u.listUserEntities()).Methods("GET")
		private.Path("/api/v1/user/entities/{entityID}").HandlerFunc(u.updateUserEntity()).Methods("PATCH")

		private.Path("/api/v1/users/removeFromFavoriteEntities").HandlerFunc(u.removeFromFavoriteEntities()).Methods("POST")
		private.Path("/api/v1/users/toggleShowRecentMatchedTags").HandlerFunc(u.toggleShowRecentMatchedTags()).Methods("POST")
		private.Path("/api/v1/users/addToFavoriteEntities").HandlerFunc(u.addToFavoriteEntities()).Methods("POST")
	})
}

func (u *userHandler) FindByID(id string) (*types.User, error) {
	objID, _ := primitive.ObjectIDFromHex(id)
	user, err := logic.User.FindByID(objID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *userHandler) FindByEntityID(id string) (*types.User, error) {
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
	type data struct {
		UserID   string `json:"userID"`
		EntityID string `json:"entityID"`
		Token    string `json:"token"`
	}
	type respond struct {
		Data data `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SignupReqBody
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			l.Logger.Info("[INFO] UserHandler.signup failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		errs := validate.SignUp(&req)
		if logic.User.UserEmailExists(req.Email) {
			errs = append(errs, errors.New("Email address is already registered."))
		}
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		entityID, err := logic.Entity.Create(&types.Entity{
			EntityName:         req.EntityName,
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

func (u *userHandler) userProfile() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.UserProfileRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := u.FindByID(r.Header.Get("userID"))
		if err != nil {
			l.Logger.Error("[ERROR] UserHandler.userProfile failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}
		api.Respond(w, r, http.StatusOK, respond{Data: &types.UserProfileRespond{
			ID:                            user.ID.Hex(),
			Email:                         user.Email,
			UserPhone:                     user.Telephone,
			FirstName:                     user.FirstName,
			LastName:                      user.LastName,
			LastLoginIP:                   user.LastLoginIP,
			LastLoginDate:                 user.LastLoginDate,
			DailyEmailMatchNotification:   user.DailyNotification,
			ShowTagsMatchedSinceLastLogin: user.ShowRecentMatchedTags,
		}})
	}
}

func (u *userHandler) updateUser() func(http.ResponseWriter, *http.Request) {
	type data struct {
		ID                            string    `json:"id"`
		Email                         string    `json:"email"`
		FirstName                     string    `json:"firstName"`
		LastName                      string    `json:"lastName"`
		UserPhone                     string    `json:"userPhone"`
		LastLoginIP                   string    `json:"lastLoginIP"`
		LastLoginDate                 time.Time `json:"lastLoginDate"`
		DailyEmailMatchNotification   *bool     `json:"dailyEmailMatchNotification"`
		ShowTagsMatchedSinceLastLogin *bool     `json:"showTagsMatchedSinceLastLogin"`
	}
	type respond struct {
		Data data `json:"data"`
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

		errs := validate.UpdateUser(&req)
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

		api.Respond(w, r, http.StatusOK, respond{Data: data{
			ID:                            user.ID.Hex(),
			Email:                         user.Email,
			UserPhone:                     user.Telephone,
			FirstName:                     user.FirstName,
			LastName:                      user.LastName,
			LastLoginIP:                   user.LastLoginIP,
			LastLoginDate:                 user.LastLoginDate,
			DailyEmailMatchNotification:   user.DailyNotification,
			ShowTagsMatchedSinceLastLogin: user.ShowRecentMatchedTags,
		}})
	}
}

func (u *userHandler) listUserEntities() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data []*types.UserEntityRespond `json:"data"`
	}
	toData := func(entities []*types.Entity) []*types.UserEntityRespond {
		result := []*types.UserEntityRespond{}
		for _, entity := range entities {
			result = append(result, &types.UserEntityRespond{
				ID:                 entity.ID.Hex(),
				EntityName:         entity.EntityName,
				EntityPhone:        entity.EntityPhone,
				IncType:            entity.IncType,
				CompanyNumber:      entity.CompanyNumber,
				Website:            entity.Website,
				Turnover:           entity.Turnover,
				Description:        entity.Description,
				LocationAddress:    entity.LocationAddress,
				LocationCity:       entity.LocationCity,
				LocationRegion:     entity.LocationRegion,
				LocationPostalCode: entity.LocationPostalCode,
				LocationCountry:    entity.LocationCountry,
				Status:             entity.Status,
				Offers:             util.TagFieldToNames(entity.Offers),
				Wants:              util.TagFieldToNames(entity.Wants),
			})
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

func isEntityBelongsToUser(entityID, userID primitive.ObjectID) bool {
	user, err := logic.User.FindByID(userID)
	if err != nil {
		return false
	}
	for _, entity := range user.Entities {
		if entity.Hex() == entityID.Hex() {
			return true
		}
	}
	return false
}

func updateTags(old *types.Entity, offers, wants []string) {
	if len(offers) == 0 && len(wants) == 0 {
		return
	}

	var offersAdded, offersRemoved, wantsAdded, wantsRemoved []string
	if len(offers) != 0 {
		offersAdded, offersRemoved = util.TagDifference(offers, util.TagFieldToNames(old.Offers))
	}
	if len(wants) != 0 {
		wantsAdded, wantsRemoved = util.TagDifference(wants, util.TagFieldToNames(old.Wants))
	}

	err := logic.Entity.UpdateTags(old.ID, &types.TagDifference{
		OffersAdded:   offersAdded,
		OffersRemoved: offersRemoved,
		WantsAdded:    wantsAdded,
		WantsRemoved:  wantsRemoved,
	})
	if err != nil {
		l.Logger.Error("[Error] UpdateTags failed:", zap.Error(err))
		return
	}

	// User Update tags logic:
	// 	1. Update the tags collection only when the entity is in accepted status.
	if util.IsAcceptedStatus(old.Status) {
		err := TagHandler.SaveOfferTags(offersAdded)
		if err != nil {
			l.Logger.Error("[Error] SaveOfferTags failed:", zap.Error(err))
		}
		err = TagHandler.SaveWantTags(wantsAdded)
		if err != nil {
			l.Logger.Error("[Error] SaveWantTags failed:", zap.Error(err))
		}
	}
}

func (u *userHandler) updateUserEntity() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.UserEntityRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateUserEntityReqBody
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			l.Logger.Info("[INFO] UserHandler.updateUserEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		errs := validate.UpdateUserEntity(&req)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		req.Offers, req.Wants = util.FormatTags(req.Offers), util.FormatTags(req.Wants)

		vars := mux.Vars(r)
		entityID, _ := primitive.ObjectIDFromHex(vars["entityID"])
		userID, _ := primitive.ObjectIDFromHex(r.Header.Get("userID"))
		if !isEntityBelongsToUser(entityID, userID) {
			api.Respond(w, r, http.StatusForbidden, api.ErrPermissionDenied)
			return
		}

		oldEntity, err := logic.Entity.FindByID(entityID)
		if err != nil {
			l.Logger.Error("[Error] UserHandler.updateUserEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		entity, err := logic.Entity.FindOneAndUpdate(&types.Entity{
			ID:                 entityID,
			EntityName:         req.EntityName,
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

		go updateTags(oldEntity, req.Offers, req.Wants)

		offers := req.Offers
		if len(offers) == 0 {
			offers = util.TagFieldToNames(entity.Offers)
		}
		wants := req.Wants
		if len(wants) == 0 {
			wants = util.TagFieldToNames(entity.Wants)
		}
		api.Respond(w, r, http.StatusOK, respond{Data: &types.UserEntityRespond{
			ID:                 entity.ID.Hex(),
			EntityName:         entity.EntityName,
			EntityPhone:        entity.EntityPhone,
			IncType:            entity.IncType,
			CompanyNumber:      entity.CompanyNumber,
			Website:            entity.Website,
			Turnover:           entity.Turnover,
			Description:        entity.Description,
			LocationAddress:    entity.LocationAddress,
			LocationCity:       entity.LocationCity,
			LocationRegion:     entity.LocationRegion,
			LocationPostalCode: entity.LocationPostalCode,
			LocationCountry:    entity.LocationCountry,
			Status:             entity.Status,
			Offers:             offers,
			Wants:              wants,
		}})
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

func (u *userHandler) addToFavoriteEntities() func(http.ResponseWriter, *http.Request) {
	type request struct {
		ID string `json:"id"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil || req.ID == "" {
			if err != nil {
				l.Logger.Error("AppServer AddToFavoriteEntities failed", zap.Error(err))
			} else {
				l.Logger.Error("AppServer AddToFavoriteEntities failed", zap.String("error", "request entity id is empty"))
			}
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}
		bID, err := primitive.ObjectIDFromHex(req.ID)
		if err != nil {
			l.Logger.Error("AppServer AddToFavoriteEntities failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}

		uID, err := primitive.ObjectIDFromHex(r.Header.Get("userID"))
		if err != nil {
			l.Logger.Error("AppServer AddToFavoriteEntities failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}

		err = logic.User.AddToFavoriteEntities(uID, bID)
		if err != nil {
			l.Logger.Error("AppServer AddToFavoriteEntities failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (u *userHandler) removeFromFavoriteEntities() func(http.ResponseWriter, *http.Request) {
	type request struct {
		ID string `json:"id"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil || req.ID == "" {
			if err != nil {
				l.Logger.Error("AppServer RemoveFromFavoriteEntities failed", zap.Error(err))
			} else {
				l.Logger.Error("AppServer RemoveFromFavoriteEntities failed", zap.String("error", "request entity id is empty"))
			}
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}
		bID, err := primitive.ObjectIDFromHex(req.ID)
		if err != nil {
			l.Logger.Error("AppServer RemoveFromFavoriteEntities failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}

		uID, err := primitive.ObjectIDFromHex(r.Header.Get("userID"))
		if err != nil {
			l.Logger.Error("AppServer RemoveFromFavoriteEntities failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}

		err = logic.User.RemoveFromFavoriteEntities(uID, bID)
		if err != nil {
			l.Logger.Error("AppServer RemoveFromFavoriteEntities failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
