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

		adminPrivate.Path("/users").HandlerFunc(handler.adminSearchUser()).Methods("GET")
		adminPrivate.Path("/users/{userID}").HandlerFunc(handler.adminGetUser()).Methods("GET")
		adminPrivate.Path("/users/{userID}").HandlerFunc(handler.adminUpdateUser()).Methods("PATCH")
		adminPrivate.Path("/users/{userID}").HandlerFunc(handler.adminDeleteUser()).Methods("DELETE")
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
		return nil, err
	}
	user, err := logic.User.FindByEntityID(objID)
	if err != nil {
		return nil, err
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

// POST /login

func (handler *userHandler) login() func(http.ResponseWriter, *http.Request) {
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

		user, err := logic.User.Login(req.Email, req.Password)
		if err != nil {

			go logic.User.IncLoginAttempts(req.Email)
			go logic.UserAction.LoginFail(req.Email, util.IPAddress(r))

			l.Logger.Info("[INFO] UserHandler.login failed", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)

			return
		}
		loginInfo, err := logic.User.UpdateLoginInfo(user.ID, util.IPAddress(r))
		if err != nil {
			l.Logger.Error("[Error] AdminUser.UpdateLoginInfo failed:", zap.Error(err))
		}

		go logic.UserAction.Login(user, util.IPAddress(r))

		token, err := util.GenerateToken(user.ID.Hex(), false)

		api.Respond(w, r, http.StatusOK, respond{Data: respondData(loginInfo, token)})
	}
}

// POST /signup

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
		req, errs := types.NewSignupReq(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		if logic.User.UserEmailExists(req.Email) {
			api.Respond(w, r, http.StatusBadRequest, errors.New("Email address is already registered."))
			return
		}

		createdEntity, err := logic.Entity.Create(&types.Entity{
			Name:             req.EntityName,
			Email:            req.Email,
			IncType:          req.IncType,
			CompanyNumber:    req.CompanyNumber,
			Telephone:        req.EntityPhone,
			Website:          req.Website,
			DeclaredTurnover: req.Turnover,
			Description:      req.Description,
			Address:          req.LocationAddress,
			City:             req.LocationCity,
			Region:           req.LocationRegion,
			PostalCode:       req.LocationPostalCode,
			Country:          req.LocationCountry,
			Offers:           types.ToTagFields(req.Offers),
			Wants:            types.ToTagFields(req.Wants),
		})
		if err != nil {
			l.Logger.Error("[ERROR] UserHandler.signup failed", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}
		createdUser, err := logic.User.Create(&types.User{
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
		err = logic.Entity.AssociateUser(createdEntity.ID, createdUser.ID)
		if err != nil {
			l.Logger.Error("[ERROR] UserHandler.signup failed", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}
		err = logic.User.AssociateEntity(createdUser.ID, createdEntity.ID)
		if err != nil {
			l.Logger.Error("[ERROR] UserHandler.signup failed", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		token, err := util.GenerateToken(createdUser.ID.Hex(), false)
		if err != nil {
			l.Logger.Error("[ERROR] UserHandler.signup failed", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		go logic.UserAction.Signup(createdUser, createdEntity)

		api.Respond(w, r, http.StatusOK, respond{Data: data{
			UserID:   createdUser.ID.Hex(),
			EntityID: createdEntity.ID.Hex(),
			Token:    token,
		}})
	}
}

// POST /logout

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
			go logic.UserAction.LostPassword(user, util.IPAddress(r))
			api.Respond(w, r, http.StatusOK, respond{Data: data{Token: uid.String()}})
			return
		}

		go logic.UserAction.LostPassword(user, util.IPAddress(r))

		api.Respond(w, r, http.StatusOK)
	}
}

func (handler *userHandler) passwordReset() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		var req types.ResetPasswordReq
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

// POST /password-change

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

		go logic.UserAction.ChangePassword(user, util.IPAddress(r))

		api.Respond(w, r, http.StatusOK)
	}
}

// GET /user

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
		api.Respond(w, r, http.StatusOK, respond{Data: types.NewUserRespond(user)})
	}
}

// PATCH /user

func (handler *userHandler) updateUser() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.UserRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := types.NewUpdateUserReq(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		originUser, err := logic.User.FindByStringID(r.Header.Get("userID"))
		if err != nil {
			l.Logger.Info("[INFO] UserHandler.updateUser failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		updated, err := logic.User.FindOneAndUpdate(originUser.ID, &types.User{
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

		go logic.UserAction.ModifyUser(originUser, updated)

		api.Respond(w, r, http.StatusOK, respond{Data: types.NewUserRespond(updated)})
	}
}

// GET /user/entities

func (handler *userHandler) listUserEntities() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data []*types.EntityRespond `json:"data"`
	}
	toData := func(entities []*types.Entity) []*types.EntityRespond {
		result := []*types.EntityRespond{}
		for _, entity := range entities {
			result = append(result, types.NewEntityRespondWithEmail(entity))
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

// PATCH /user/entities/{entityID}

func (handler *userHandler) updateUserEntity() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.EntityRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := handler.newUpdateUserEntityReq(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		if !handler.IsEntityBelongsToUser(req.OriginEntity.ID.Hex(), r.Header.Get("userID")) {
			api.Respond(w, r, http.StatusForbidden, api.ErrPermissionDenied)
			return
		}

		updated, err := logic.Entity.FindOneAndUpdate(req)
		if err != nil {
			l.Logger.Info("[INFO] UserHandler.updateUserEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		go EntityHandler.UpdateOfferAndWants(&types.UpdateOfferAndWants{
			EntityID:      req.OriginEntity.ID,
			OriginStatus:  req.OriginEntity.Status,
			UpdatedStatus: updated.Status,
			UpdatedOffers: types.TagFieldToNames(updated.Offers),
			UpdatedWants:  types.TagFieldToNames(updated.Wants),
			AddedOffers:   req.AddedOffers,
			AddedWants:    req.AddedWants,
		})

		go logic.UserAction.ModifyEntity(r.Header.Get("userID"), req.OriginEntity, updated)

		api.Respond(w, r, http.StatusOK, respond{Data: types.NewEntityRespondWithEmail(updated)})
	}
}

func (handler *userHandler) newUpdateUserEntityReq(r *http.Request) (*types.UpdateUserEntityReq, []error) {
	originEntity, err := logic.Entity.FindByStringID(mux.Vars(r)["entityID"])
	if err != nil {
		return nil, []error{err}
	}

	var j types.UpdateUserEntityJSON
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&j)
	if err != nil {
		return nil, []error{err}
	}

	return types.NewUpdateUserEntityReq(j, originEntity)
}

// Admin

func (handler *userHandler) adminGetUser() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.AdminGetUserRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := types.NewAdminGetUserReq(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		user, err := logic.User.FindByID(req.UserID)
		if err != nil {
			l.Logger.Error("[Error] UserHandler.getUser failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		entities, err := logic.Entity.FindByIDs(user.Entities)
		if err != nil {
			l.Logger.Error("[Error] UserHandler.getUser failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		api.Respond(w, r, http.StatusOK, respond{Data: types.NewAdminGetUserRespond(user, entities)})
	}
}

func (handler *userHandler) adminSearchUser() func(http.ResponseWriter, *http.Request) {
	type meta struct {
		NumberOfResults int `json:"numberOfResults"`
		TotalPages      int `json:"totalPages"`
	}
	type respond struct {
		Data []*types.AdminGetUserRespond `json:"data"`
		Meta meta                         `json:"meta"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := types.NewAdminSearchUserReq(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		searchUserResult, err := logic.User.AdminSearchUser(req)
		if err != nil {
			l.Logger.Error("[Error] UserHandler.adminSearchUser failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		res, err := handler.newAdminSearchUserRespond(searchUserResult)
		if err != nil {
			l.Logger.Error("[Error] UserHandler.adminSearchUser failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		api.Respond(w, r, http.StatusOK, respond{
			Data: res,
			Meta: meta{
				TotalPages:      searchUserResult.TotalPages,
				NumberOfResults: searchUserResult.NumberOfResults,
			},
		})
	}
}

func (handler *userHandler) newAdminSearchUserRespond(searchUserResult *types.SearchUserResult) ([]*types.AdminGetUserRespond, error) {
	respond := []*types.AdminGetUserRespond{}
	for _, user := range searchUserResult.Users {
		entities, err := logic.Entity.FindByIDs(user.Entities)
		if err != nil {
			return nil, err
		}
		respond = append(respond, types.NewAdminGetUserRespond(user, entities))
	}
	return respond, nil
}

// PATCH /admin/users/{userID}

func (handler *userHandler) adminUpdateUser() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.AdminGetUserRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := handler.newAdminUpdateUserReq(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		updated, err := logic.User.AdminFindOneAndUpdate(req)
		if err != nil {
			l.Logger.Error("[Error] UserHandler.adminUpdateUser failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		entities, err := logic.Entity.FindByIDs(updated.Entities)
		if err != nil {
			l.Logger.Error("[Error] UserHandler.adminUpdateUser failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		go logic.UserAction.AdminModifyUser(r.Header.Get("userID"), req.OriginUser, updated)

		api.Respond(w, r, http.StatusOK, respond{Data: types.NewAdminGetUserRespond(updated, entities)})
	}
}

func (handler *userHandler) newAdminUpdateUserReq(r *http.Request) (*types.AdminUpdateUserReq, []error) {
	originUser, err := logic.User.FindByStringID(mux.Vars(r)["userID"])
	if err != nil {
		return nil, []error{err}
	}

	var j types.AdminUpdateUserJSON
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&j)
	if err != nil {
		return nil, []error{err}
	}

	if j.Entities != nil {
		for _, entityID := range *j.Entities {
			_, err := logic.Entity.FindByStringID(entityID)
			if err != nil {
				return nil, []error{err}
			}
		}
	}

	return types.NewAdminUpdateUserReq(j, originUser)
}

// DELETE /admin/users/{userID}

func (handler *userHandler) adminDeleteUser() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.AdminDeleteUserRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := types.NewAdminDeleteUserReq(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		deleted, err := logic.User.AdminFindOneAndDelete(req.UserID)
		if err != nil {
			l.Logger.Error("[Error] UserHandler.adminDeleteUser failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		go logic.UserAction.AdminDeleteUser(r.Header.Get("userID"), deleted)

		api.Respond(w, r, http.StatusOK, respond{Data: types.NewAdminDeleteUserRespond(deleted)})
	}
}
