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
	"github.com/ic3network/mccs-alpha-api/internal/pkg/l"
	"github.com/ic3network/mccs-alpha-api/util"
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

		private.Path("/entities/search/match-tags").HandlerFunc(handler.searhMatchTags()).Methods("GET")

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

func (handler *entityHandler) UpdateOffersAndWants(old *types.Entity, offers, wants []string) {
	if len(offers) == 0 && len(wants) == 0 {
		return
	}

	tagDifference := types.NewTagDifference(types.TagFieldToNames(old.Offers), offers, types.TagFieldToNames(old.Wants), wants)
	err := logic.Entity.UpdateTags(old.ID, tagDifference)
	if err != nil {
		l.Logger.Error("[Error] EntityHandler.UpdateOffersAndWants failed:", zap.Error(err))
		return
	}

	if util.IsAcceptedStatus(old.Status) {
		// User Update tags logic:
		// 	1. Update the tags collection only when the entity is in accepted status.
		err := TagHandler.UpdateOffers(tagDifference.NewAddedOffers)
		if err != nil {
			l.Logger.Error("[Error] EntityHandler.UpdateOffersAndWants failed:", zap.Error(err))
		}
		err = TagHandler.UpdateWants(tagDifference.NewAddedWants)
		if err != nil {
			l.Logger.Error("[Error] EntityHandler.UpdateOffersAndWants failed:", zap.Error(err))
		}
	}
}

func (handler *entityHandler) getSearchEntityQueryParams(q url.Values) (*types.SearchEntityReqBody, error) {
	query, err := types.NewSearchEntityReqBody(q)
	if err != nil {
		return nil, err
	}
	query.FavoriteEntities = handler.getFavoriteEntities(q.Get("querying_entity_id"))
	return query, nil
}

func (handler *entityHandler) getFavoriteEntities(entityID string) []primitive.ObjectID {
	entity, err := EntityHandler.FindByID(entityID)
	if err == nil {
		return entity.FavoriteEntities
	}
	return []primitive.ObjectID{}
}

func (handler *entityHandler) getQueryingEntityStatus(entityID string) string {
	entity, err := EntityHandler.FindByID(entityID)
	if err == nil {
		return entity.Status
	}
	return ""
}

func (handler *entityHandler) searchEntity() func(http.ResponseWriter, *http.Request) {
	type meta struct {
		NumberOfResults int `json:"numberOfResults"`
		TotalPages      int `json:"totalPages"`
	}
	type respond struct {
		Data []*types.SearchEntityRespond `json:"data"`
		Meta meta                         `json:"meta"`
	}
	toData := func(query *types.SearchEntityReqBody, entities []*types.Entity) []*types.SearchEntityRespond {
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

func (handler *entityHandler) getEntity() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.SearchEntityRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		searchEntity, err := logic.Entity.FindByStringID(vars["searchEntityID"])
		if err != nil {
			l.Logger.Info("[INFO] EntityHandler.getEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		q := r.URL.Query()
		queryingEntityID := q.Get("querying_entity_id")

		if queryingEntityID != "" {
			if r.Header.Get("userID") == "" {
				api.Respond(w, r, http.StatusUnauthorized, api.ErrUnauthorized)
				return
			}
			if !UserHandler.IsEntityBelongsToUser(queryingEntityID, r.Header.Get("userID")) {
				api.Respond(w, r, http.StatusForbidden, api.ErrPermissionDenied)
				return
			}
		}

		queryingEntityStatus := handler.getQueryingEntityStatus(queryingEntityID)
		favoriteEntities := handler.getFavoriteEntities(queryingEntityID)

		api.Respond(w, r, http.StatusOK, respond{Data: types.NewSearchEntityRespond(searchEntity, queryingEntityStatus, favoriteEntities)})
	}
}

func (handler *entityHandler) addToFavoriteEntities() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AddToFavoriteReqBody
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			l.Logger.Info("[Info] EntityHandler.addToFavorite failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		errs := req.Validate()
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		err = logic.Entity.AddToFavoriteEntities(&req)
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

func (handler *entityHandler) sendEmailToEntity() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := types.NewEmailReqBody(r)
		if err != nil {
			l.Logger.Info("[Info] EntityHandler.sendEmailToEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		errs := req.Validate()
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

		err = email.SendContactEntity(ReceiverEntity.EntityName, ReceiverEntity.Email, SenderEntity.EntityName, SenderEntity.Email, req.Body)
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
				SenderEntityName:   SenderEntity.EntityName,
				ReceiverEntityName: ReceiverEntity.EntityName,
				Body:               req.Body,
			}})
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// Admin

func (handler *entityHandler) adminSearchEntity() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data []*types.AdminGetEntityRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := types.NewAdminSearchEntityReqBody(r)
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

		api.Respond(w, r, http.StatusOK, respond{Data: res})
	}
}

func (handler *entityHandler) newAdminSearchEntityRespond(searchEntityResult *types.SearchEntityResult) ([]*types.AdminGetEntityRespond, error) {
	respond := []*types.AdminGetEntityRespond{}
	for _, entity := range searchEntityResult.Entities {
		users, err := logic.User.FindByIDs(entity.Users)
		if err != nil {
			return nil, err
		}
		respond = append(respond, types.NewAdminGetEntityRespond(entity, users))
	}
	return respond, nil
}

func (handler *entityHandler) adminGetEntity() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.AdminGetEntityRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := types.NewAdminGetEntityReqBody(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		entity, err := logic.Entity.FindByID(req.EntityID)
		if err != nil {
			l.Logger.Error("[Error] EntityHandler.adminGetEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		users, err := logic.User.FindByIDs(entity.Users)
		if err != nil {
			l.Logger.Error("[Error] EntityHandler.adminGetEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		api.Respond(w, r, http.StatusOK, respond{Data: types.NewAdminGetEntityRespond(entity, users)})
	}
}

func (handler *entityHandler) adminUpdateEntity() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.AdminEntityRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := types.NewAdminUpdateEntityReqBody(r)
		if err != nil {
			l.Logger.Info("[INFO] AdminEntityHandler.updateEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		errs := req.Validate()
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		oldEntity, err := logic.Entity.FindByID(req.EntityID)
		if err != nil {
			l.Logger.Error("[Error] AdminEntityHandler.updateEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		newEntity, err := logic.Entity.AdminFindOneAndUpdate(&types.Entity{
			ID:                 req.EntityID,
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
			Categories:         req.Categories,
			Status:             req.Status,
		})
		if err != nil {
			l.Logger.Error("[Error] AdminEntityHandler.updateEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		// Update offers and wants are seperate processes.
		go handler.updateOfferAndWants(oldEntity, req.Status, req.Offers, req.Wants)
		go handler.updateEntityMemberStartedAt(oldEntity, req.Status)
		go CategoryHandler.Update(req.Categories)

		if len(req.Offers) != 0 {
			newEntity.Offers = types.ToTagFields(req.Offers)
		}
		if len(req.Wants) != 0 {
			newEntity.Wants = types.ToTagFields(req.Wants)
		}
		api.Respond(w, r, http.StatusOK, respond{Data: types.NewAdminEntityRespond(newEntity)})
	}
}

func (handler *entityHandler) updateEntityMemberStartedAt(oldEntity *types.Entity, newStatus string) {
	// Set timestamp when first trading status applied.
	if oldEntity.MemberStartedAt.IsZero() && (oldEntity.Status == constant.Entity.Accepted) && (newStatus == constant.Trading.Accepted) {
		logic.Entity.SetMemberStartedAt(oldEntity.ID)
	}
}

func (handler *entityHandler) updateOfferAndWants(oldEntity *types.Entity, newStatus string, offers []string, wants []string) {
	tagDifference := types.NewTagDifference(types.TagFieldToNames(oldEntity.Offers), offers, types.TagFieldToNames(oldEntity.Wants), wants)
	err := logic.Entity.UpdateTags(oldEntity.ID, tagDifference)
	if err != nil {
		l.Logger.Error("[Error] AdminEntityHandler.updateOfferAndWants failed:", zap.Error(err))
		return
	}

	// Admin Update tags logic:
	// 	1. When a entity' status is changed from pending/rejected to accepted.
	// 	   - update all tags.
	// 	2. When the entity is in accepted status.
	//	   - only update added tags.
	if !util.IsAcceptedStatus(oldEntity.Status) && util.IsAcceptedStatus(newStatus) {
		err := logic.Entity.UpdateAllTagsCreatedAt(oldEntity.ID, time.Now())
		if err != nil {
			l.Logger.Error("[Error] AdminEntityHandler.updateOfferAndWants failed:", zap.Error(err))
		}
		err = TagHandler.UpdateOffers(tagDifference.Offers)
		if err != nil {
			l.Logger.Error("[Error] AdminEntityHandler.updateOfferAndWants failed:", zap.Error(err))
		}
		err = TagHandler.UpdateWants(tagDifference.Wants)
		if err != nil {
			l.Logger.Error("[Error] AdminEntityHandler.updateOfferAndWants failed:", zap.Error(err))
		}
	}
	if util.IsAcceptedStatus(oldEntity.Status) && util.IsAcceptedStatus(newStatus) {
		err := TagHandler.UpdateOffers(tagDifference.NewAddedOffers)
		if err != nil {
			l.Logger.Error("[Error] AdminEntityHandler.updateOfferAndWants failed:", zap.Error(err))
		}
		err = TagHandler.UpdateWants(tagDifference.NewAddedWants)
		if err != nil {
			l.Logger.Error("[Error] AdminEntityHandler.updateOfferAndWants failed:", zap.Error(err))
		}
	}
}

func (handler *entityHandler) adminDeleteEntity() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.AdminGetEntityRespond `json:"data"`
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

		users, err := logic.User.FindByIDs(deleted.Users)
		if err != nil {
			l.Logger.Error("[Error] EntityHandler.adminDeleteEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		api.Respond(w, r, http.StatusOK, respond{Data: types.NewAdminGetEntityRespond(deleted, users)})
	}
}

// TO BE REMOVED

func (handler *entityHandler) searhMatchTags() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/entities/search?"+r.URL.Query().Encode(), http.StatusFound)
	}
}

func (handler *entityHandler) entityStatus() func(http.ResponseWriter, *http.Request) {
	type response struct {
		Status string `json:"status"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var entity *types.Entity
		var err error

		q := r.URL.Query()

		if q.Get("entity_id") != "" {
			objID, err := primitive.ObjectIDFromHex(q.Get("entity_id"))
			if err != nil {
				l.Logger.Error("EntityHandler.entityStatus failed", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			entity, err = logic.Entity.FindByID(objID)
			if err != nil {
				l.Logger.Error("EntityHandler.entityStatus failed", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			entity, err = EntityHandler.FindByUserID(r.Header.Get("userID"))
			if err != nil {
				l.Logger.Error("EntityHandler.entityStatus failed", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		res := &response{Status: entity.Status}
		js, err := json.Marshal(res)
		if err != nil {
			l.Logger.Error("EntityHandler.entityStatus failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

func (handler *entityHandler) getEntityName() func(http.ResponseWriter, *http.Request) {
	type response struct {
		Name string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		user, err := logic.User.FindByEmail(q.Get("email"))
		if err != nil {
			l.Logger.Error("EntityHandler.getEntityName failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		entity, err := logic.Entity.FindByID(user.Entities[0])
		if err != nil {
			l.Logger.Error("EntityHandler.getEntityName failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		res := response{Name: entity.EntityName}
		js, err := json.Marshal(res)
		if err != nil {
			l.Logger.Error("EntityHandler.getEntityName failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

func (handler *entityHandler) tradingMemberStatus() func(http.ResponseWriter, *http.Request) {
	type response struct {
		Self  bool `json:"self"`
		Other bool `json:"other"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		objID, err := primitive.ObjectIDFromHex(q.Get("entity_id"))
		if err != nil {
			l.Logger.Error("EntityHandler.tradingMemberStatus failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		other, err := logic.Entity.FindByID(objID)
		if err != nil {
			l.Logger.Error("EntityHandler.tradingMemberStatus failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		self, err := EntityHandler.FindByUserID(r.Header.Get("userID"))
		if err != nil {
			l.Logger.Error("EntityHandler.tradingMemberStatus failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		res := &response{}
		if self.Status == constant.Trading.Accepted {
			res.Self = true
		}
		if other.Status == constant.Trading.Accepted {
			res.Other = true
		}
		js, err := json.Marshal(res)
		if err != nil {
			l.Logger.Error("EntityHandler.tradingMemberStatus failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}
