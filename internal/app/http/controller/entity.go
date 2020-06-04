package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/global/constant"
	"github.com/ic3network/mccs-alpha-api/internal/app/api"
	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/email"
	"github.com/ic3network/mccs-alpha-api/util"
	"github.com/ic3network/mccs-alpha-api/util/l"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type entityHandler struct {
	once *sync.Once
}

var EntityHandler = newEntityHandler()

func newEntityHandler() *entityHandler {
	return &entityHandler{
		once: new(sync.Once),
	}
}

func (handler *entityHandler) RegisterRoutes(
	public *mux.Router,
	private *mux.Router,
	adminPublic *mux.Router,
	adminPrivate *mux.Router,
) {
	handler.once.Do(func() {
		public.Path("/entities").HandlerFunc(handler.searchEntity()).Methods("GET")
		public.Path("/entities/{searchEntityID}").HandlerFunc(handler.getEntity()).Methods("GET")
		private.Path("/favorites").HandlerFunc(handler.addToFavoriteEntities()).Methods("POST")
		private.Path("/send-email").HandlerFunc(handler.sendEmailToEntity()).Methods("POST")
		private.Path("/balance").HandlerFunc(handler.getBalance()).Methods("GET")

		adminPrivate.Path("/entities").HandlerFunc(handler.adminSearchEntity()).Methods("GET")
		adminPrivate.Path("/entities/{entityID}").HandlerFunc(handler.adminGetEntity()).Methods("GET")
		adminPrivate.Path("/entities/{entityID}").HandlerFunc(handler.adminUpdateEntity()).Methods("PATCH")
		adminPrivate.Path("/entities/{entityID}").HandlerFunc(handler.adminDeleteEntity()).Methods("DELETE")
	})
}

func (handler *entityHandler) FindByID(entityID string) (*types.Entity, error) {
	objID, err := primitive.ObjectIDFromHex(entityID)
	if err != nil {
		return nil, err
	}
	entity, err := logic.Entity.FindByID(objID)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (handler *entityHandler) FindByEmail(email string) (*types.Entity, error) {
	user, err := logic.User.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	bs, err := logic.Entity.FindByID(user.Entities[0])
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func (handler *entityHandler) FindByUserID(uID string) (*types.Entity, error) {
	user, err := UserHandler.FindByID(uID)
	if err != nil {
		return nil, err
	}
	bs, err := logic.Entity.FindByID(user.Entities[0])
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func (handler *entityHandler) getFavoriteEntities(entityID string) []primitive.ObjectID {
	entity, err := EntityHandler.FindByID(entityID)
	if err == nil {
		return entity.FavoriteEntities
	}
	return []primitive.ObjectID{}
}

// GET /entities

func (handler *entityHandler) searchEntity() func(http.ResponseWriter, *http.Request) {
	type meta struct {
		NumberOfResults int `json:"numberOfResults"`
		TotalPages      int `json:"totalPages"`
	}
	type respond struct {
		Data []*types.SearchEntityRespond `json:"data"`
		Meta meta                         `json:"meta"`
	}
	toData := func(query *types.SearchEntityReq, entities []*types.Entity) []*types.SearchEntityRespond {
		result := []*types.SearchEntityRespond{}
		queryingEntityStatus := handler.getQueryingEntityStatus(query.QueryingEntityID)
		for _, entity := range entities {
			result = append(result, types.NewSearchEntityRespond(entity, queryingEntityStatus, query.FavoriteEntities))
		}
		return result
	}
	return func(w http.ResponseWriter, r *http.Request) {
		query, err := handler.getSearchEntityQueryParams(r.URL.Query())
		if err != nil {
			l.Logger.Info("[INFO] EntityHandler.searchEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		errs := query.Validate()
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		if query.QueryingEntityID != "" {
			if r.Header.Get("userID") == "" {
				api.Respond(w, r, http.StatusUnauthorized, api.ErrUnauthorized)
				return
			}
			if !UserHandler.IsEntityBelongsToUser(query.QueryingEntityID, r.Header.Get("userID")) {
				api.Respond(w, r, http.StatusForbidden, api.ErrPermissionDenied)
				return
			}
		}

		found, err := logic.Entity.Search(query)
		if err != nil {
			l.Logger.Error("[Error] EntityHandler.searchEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		api.Respond(w, r, http.StatusOK, respond{
			Data: toData(query, found.Entities),
			Meta: meta{
				TotalPages:      found.TotalPages,
				NumberOfResults: found.NumberOfResults,
			},
		})
	}
}

func (handler *entityHandler) getSearchEntityQueryParams(q url.Values) (*types.SearchEntityReq, error) {
	query, err := types.NewSearchEntityReq(q)
	if err != nil {
		return nil, err
	}
	query.FavoriteEntities = handler.getFavoriteEntities(q.Get("querying_entity_id"))
	return query, nil
}

func (handler *entityHandler) getQueryingEntityStatus(entityID string) string {
	entity, err := EntityHandler.FindByID(entityID)
	if err == nil {
		return entity.Status
	}
	return ""
}

// GET /entities/{entityID}

func (handler *entityHandler) getEntity() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.SearchEntityRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := types.NewGetEntityReq(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		searchEntity, err := logic.Entity.FindByStringID(req.SearchEntityID)
		if err != nil {
			l.Logger.Info("[INFO] EntityHandler.getEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}
		if req.QueryingEntityID != "" {
			if r.Header.Get("userID") == "" {
				api.Respond(w, r, http.StatusUnauthorized, api.ErrUnauthorized)
				return
			}
			if !UserHandler.IsEntityBelongsToUser(req.QueryingEntityID, r.Header.Get("userID")) {
				api.Respond(w, r, http.StatusForbidden, api.ErrPermissionDenied)
				return
			}
		}

		queryingEntityStatus := handler.getQueryingEntityStatus(req.QueryingEntityID)
		favoriteEntities := handler.getFavoriteEntities(req.QueryingEntityID)

		api.Respond(w, r, http.StatusOK, respond{Data: types.NewSearchEntityRespond(searchEntity, queryingEntityStatus, favoriteEntities)})
	}
}

// POST favorites

func (handler *entityHandler) addToFavoriteEntities() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := types.NewAddToFavoriteReq(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		if !UserHandler.IsEntityBelongsToUser(req.AddToEntityID, r.Header.Get("userID")) {
			api.Respond(w, r, http.StatusForbidden, api.ErrPermissionDenied)
			return
		}

		err := logic.Entity.AddToFavoriteEntities(req)
		if err != nil {
			l.Logger.Error("[Error] EntityHandler.addToFavorite failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		api.Respond(w, r, http.StatusOK, nil)
	}
}

func (handler *entityHandler) checkEntityStatus(SenderEntity, ReceiverEntity *types.Entity) error {
	if !util.IsAcceptedStatus(SenderEntity.Status) {
		return errors.New("Sender does not have the correct status.")

	}
	if !util.IsAcceptedStatus(ReceiverEntity.Status) {
		return errors.New("Receiver does not have the correct status.")
	}
	return nil
}

// POST /send-email

func (handler *entityHandler) sendEmailToEntity() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := types.NewEmailReq(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		SenderEntity, err := handler.FindByID(req.SenderEntityID)
		if err != nil {
			l.Logger.Error("[Error] EntityHandler.sendEmailToEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}
		ReceiverEntity, err := handler.FindByID(req.ReceiverEntityID)
		if err != nil {
			l.Logger.Error("[Error] EntityHandler.sendEmailToEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}
		if !UserHandler.IsEntityBelongsToUser(req.SenderEntityID, r.Header.Get("userID")) {
			api.Respond(w, r, http.StatusForbidden, api.ErrPermissionDenied)
			return
		}
		err = handler.checkEntityStatus(SenderEntity, ReceiverEntity)
		if err != nil {
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		err = email.SendContactEntity(ReceiverEntity.Name, ReceiverEntity.Email, SenderEntity.Name, SenderEntity.Email, req.Body)
		if err != nil {
			l.Logger.Error("[Error] EntityHandler.sendEmailToEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		if viper.GetString("env") == "development" {
			type data struct {
				SenderEntityName   string `json:"sender_entity_name"`
				ReceiverEntityName string `json:"receiver_entity_name"`
				Body               string `json:"body"`
			}
			type respond struct {
				Data data `json:"data"`
			}
			api.Respond(w, r, http.StatusOK, respond{Data: data{
				SenderEntityName:   SenderEntity.Name,
				ReceiverEntityName: ReceiverEntity.Name,
				Body:               req.Body,
			}})
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// GET /balance

func (handler *entityHandler) getBalance() func(http.ResponseWriter, *http.Request) {
	type data struct {
		Unit    string  `json:"unit"`
		Balance float64 `json:"balance"`
	}
	type respond struct {
		Data data `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		query, errs := types.NewBalanceQuery(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		if !UserHandler.IsEntityBelongsToUser(query.QueryingEntityID, r.Header.Get("userID")) {
			api.Respond(w, r, http.StatusForbidden, api.ErrPermissionDenied)
			return
		}

		account, err := logic.Account.FindByEntityID(query.QueryingEntityID)
		if err != nil {
			l.Logger.Error("[Error] EntityHandler.getBalance failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		api.Respond(w, r, http.StatusOK, respond{Data: data{
			Unit:    constant.Unit.UK,
			Balance: account.Balance,
		}})
	}
}

// GET /admin/entities

func (handler *entityHandler) adminSearchEntity() func(http.ResponseWriter, *http.Request) {
	type meta struct {
		NumberOfResults int `json:"numberOfResults"`
		TotalPages      int `json:"totalPages"`
	}
	type respond struct {
		Data []*types.AdminSearchEntityRespond `json:"data"`
		Meta meta                              `json:"meta"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := types.NewAdminSearchEntityReq(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		searchEntityResult, err := logic.Entity.AdminSearch(req)
		if err != nil {
			l.Logger.Error("[Error] EntityHandler.adminSearchEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		res, err := handler.newAdminSearchEntityRespond(searchEntityResult)
		if err != nil {
			l.Logger.Error("[Error] EntityHandler.adminSearchEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		api.Respond(w, r, http.StatusOK, respond{
			Data: res,
			Meta: meta{
				TotalPages:      searchEntityResult.TotalPages,
				NumberOfResults: searchEntityResult.NumberOfResults,
			},
		})
	}
}

func (handler *entityHandler) newAdminSearchEntityRespond(searchEntityResult *types.SearchEntityResult) ([]*types.AdminSearchEntityRespond, error) {
	respond := []*types.AdminSearchEntityRespond{}
	for _, entity := range searchEntityResult.Entities {
		users, err := logic.User.FindByIDs(entity.Users)
		if err != nil {
			return nil, err
		}
		account, err := logic.Account.FindByAccountNumber(entity.AccountNumber)
		if err != nil {
			return nil, err
		}
		balanceLimit, err := logic.BalanceLimit.FindByAccountNumber(entity.AccountNumber)
		if err != nil {
			return nil, err
		}
		respond = append(respond, types.NewAdminSearchEntityRespond(entity, users, account, balanceLimit))
	}
	return respond, nil
}

// GET /admin/entities/{entityID}

func (handler *entityHandler) adminGetEntity() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.AdminGetEntityRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := types.NewAdminGetEntityReq(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		entity, err := logic.Entity.FindByStringID(req.EntityID)
		if err != nil {
			l.Logger.Error("[Error] EntityHandler.adminGetEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		res, err := handler.newAdminGetEntityRespond(entity)
		if err != nil {
			l.Logger.Error("[Error] EntityHandler.adminGetEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		api.Respond(w, r, http.StatusOK, respond{Data: res})
	}
}

func (handler *entityHandler) newAdminGetEntityRespond(entity *types.Entity) (*types.AdminGetEntityRespond, error) {
	users, err := logic.User.FindByIDs(entity.Users)
	if err != nil {
		return nil, err
	}
	account, err := logic.Account.FindByAccountNumber(entity.AccountNumber)
	if err != nil {
		return nil, err
	}
	balanceLimit, err := logic.BalanceLimit.FindByAccountNumber(entity.AccountNumber)
	if err != nil {
		return nil, err
	}
	pendingTransfers, err := logic.Transfer.AdminGetPendingTransfers(entity.AccountNumber)
	if err != nil {
		return nil, err
	}
	return types.NewAdminGetEntityRespond(entity, users, account, balanceLimit, pendingTransfers), nil
}

// PATCH /admin/entities/{entityID}

func (handler *entityHandler) adminUpdateEntity() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.AdminUpdateEntityRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := handler.newAdminUpdateEntityReq(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		updated, err := logic.Entity.AdminFindOneAndUpdate(req)
		if err != nil {
			l.Logger.Error("[Error] EntityHandler.updateEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		go handler.UpdateOfferAndWants(&types.UpdateOfferAndWants{
			EntityID:      req.OriginEntity.ID,
			OriginStatus:  req.OriginEntity.Status,
			UpdatedStatus: updated.Status,
			UpdatedOffers: types.TagFieldToNames(updated.Offers),
			UpdatedWants:  types.TagFieldToNames(updated.Wants),
			AddedOffers:   req.AddedOffers,
			AddedWants:    req.AddedWants,
		})
		go handler.updateEntityMemberStartedAt(req.OriginEntity, req.Status)
		if req.Categories != nil {
			go CategoryHandler.Update(*req.Categories)
		}

		res, err := handler.newAdminUpdateEntityRespond(req, updated)
		if err != nil {
			l.Logger.Error("[Error] EntityHandler.updateEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		go logic.UserAction.AdminModifyEntity(r.Header.Get("userID"), req.OriginEntity, updated)
		go logic.UserAction.AdminModifyBalance(r.Header.Get("userID"), req.OriginBalanceLimit, res.BalanceLimit)

		api.Respond(w, r, http.StatusOK, respond{Data: res})
	}
}

func (handler *entityHandler) newAdminUpdateEntityReq(r *http.Request) (*types.AdminUpdateEntityReq, []error) {
	originEntity, err := logic.Entity.FindByStringID(mux.Vars(r)["entityID"])
	if err != nil {
		return nil, []error{err}
	}
	originBalanceLimit, err := logic.BalanceLimit.FindByAccountNumber(originEntity.AccountNumber)
	if err != nil {
		return nil, []error{err}
	}

	var j types.AdminUpdateEntityJSON
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&j)
	if err != nil {
		return nil, []error{err}
	}

	if j.Users != nil {
		for _, userID := range *j.Users {
			_, err := logic.User.FindByStringID(userID)
			if err != nil {
				return nil, []error{err}
			}
		}
	}

	return types.NewAdminUpdateEntityReq(j, originEntity, originBalanceLimit)
}

func (handler *entityHandler) newAdminUpdateEntityRespond(req *types.AdminUpdateEntityReq, entity *types.Entity) (*types.AdminUpdateEntityRespond, error) {
	users, err := logic.User.FindByIDs(entity.Users)
	if err != nil {
		return nil, err
	}
	balanceLimit, err := logic.BalanceLimit.FindByAccountNumber(entity.AccountNumber)
	if err != nil {
		return nil, err
	}
	return types.NewAdminUpdateEntityRespond(users, entity, balanceLimit), nil
}

func (handler *entityHandler) updateEntityMemberStartedAt(oldEntity *types.Entity, newStatus string) {
	// Set timestamp when first trading status applied.
	if oldEntity.MemberStartedAt.IsZero() && (oldEntity.Status == constant.Entity.Accepted) && (newStatus == constant.Trading.Accepted) {
		logic.Entity.SetMemberStartedAt(oldEntity.ID)
	}
}

// PATCH /admin/entities/{entityID}
// PATCH /user/entities/{entityID}

func (handler *entityHandler) UpdateOfferAndWants(req *types.UpdateOfferAndWants) {
	// Update tags logic:
	// 	1. When a entity' status is changed from pending/rejected to accepted.
	// 	   - update all tags.
	// 	2. When the entity is in accepted status.
	//	   - only update added tags.
	if !util.IsAcceptedStatus(req.OriginStatus) && util.IsAcceptedStatus(req.UpdatedStatus) {
		err := logic.Entity.UpdateAllTagsCreatedAt(req.EntityID, time.Now())
		if err != nil {
			l.Logger.Error("[Error] EntityHandler.updateOfferAndWants failed:", zap.Error(err))
		}
		err = TagHandler.UpdateOffers(req.UpdatedOffers)
		if err != nil {
			l.Logger.Error("[Error] EntityHandler.updateOfferAndWants failed:", zap.Error(err))
		}
		err = TagHandler.UpdateOffers(req.UpdatedWants)
		if err != nil {
			l.Logger.Error("[Error] EntityHandler.updateOfferAndWants failed:", zap.Error(err))
		}
	}
	if util.IsAcceptedStatus(req.OriginStatus) && util.IsAcceptedStatus(req.UpdatedStatus) {
		err := TagHandler.UpdateOffers(req.AddedOffers)
		if err != nil {
			l.Logger.Error("[Error] EntityHandler.updateOfferAndWants failed:", zap.Error(err))
		}
		err = TagHandler.UpdateWants(req.AddedWants)
		if err != nil {
			l.Logger.Error("[Error] EntityHandler.updateOfferAndWants failed:", zap.Error(err))
		}
	}
}

// DELETE /admin/entities/{entityID}

func (handler *entityHandler) adminDeleteEntity() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.AdminDeleteEntityRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := types.NewAdminDeleteEntity(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		deleted, err := logic.Entity.AdminFindOneAndDelete(req.EntityID)
		if err != nil {
			l.Logger.Error("[Error] EntityHandler.adminDeleteEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		go logic.UserAction.AdminDeleteEntity(r.Header.Get("userID"), deleted)

		api.Respond(w, r, http.StatusOK, respond{Data: types.NewAdminDeleteEntityRespond(deleted)})
	}
}
