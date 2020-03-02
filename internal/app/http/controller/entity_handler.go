package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sync"

	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/global/constant"
	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/api"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/email"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/l"
	"github.com/ic3network/mccs-alpha-api/util"
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

func (b *entityHandler) RegisterRoutes(
	public *mux.Router,
	private *mux.Router,
	adminPublic *mux.Router,
	adminPrivate *mux.Router,
) {
	b.once.Do(func() {
		public.Path("/api/v1/entities").HandlerFunc(b.searchEntity()).Methods("GET")
		public.Path("/api/v1/entities/{entityID}").HandlerFunc(b.getEntity()).Methods("GET")
		private.Path("/api/v1/favorites").HandlerFunc(b.addToFavoriteEntities()).Methods("POST")
		private.Path("/api/v1/send-email").HandlerFunc(b.sendEmailToEntity()).Methods("POST")

		private.Path("/entities/search/match-tags").HandlerFunc(b.searhMatchTags()).Methods("GET")
		private.Path("/api/entityStatus").HandlerFunc(b.entityStatus()).Methods("GET")
		private.Path("/api/getEntityName").HandlerFunc(b.getEntityName()).Methods("GET")
		private.Path("/api/tradingMemberStatus").HandlerFunc(b.tradingMemberStatus()).Methods("GET")
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

func getSearchEntityQueryParams(q url.Values) (*types.SearchEntityQuery, error) {
	query, err := types.NewSearchEntityQuery(q)
	if err != nil {
		return nil, err
	}
	query.FavoriteEntities = getFavoriteEntities(q.Get("querying_entity_id"))
	return query, nil
}

func getFavoriteEntities(entityID string) []primitive.ObjectID {
	entity, err := EntityHandler.FindByID(entityID)
	if err == nil {
		return entity.FavoriteEntities
	}
	return []primitive.ObjectID{}
}

func getQueryingEntityState(entityID string) string {
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
		Data []*types.EntityRespond `json:"data"`
		Meta meta                   `json:"meta"`
	}
	toData := func(query *types.SearchEntityQuery, entities []*types.Entity) []*types.EntityRespond {
		result := []*types.EntityRespond{}
		queryingEntityState := getQueryingEntityState(query.QueryingEntityID)
		for _, entity := range entities {
			var respond *types.EntityRespond
			if util.IsTradingAccepted(queryingEntityState) && util.IsTradingAccepted(entity.Status) {
				respond = types.NewEntityRespondWithEmail(entity)
			} else {
				respond = types.NewEntityRespondWithoutEmail(entity)
			}
			respond.IsFavorite = util.ContainID(query.FavoriteEntities, entity.ID)
			result = append(result, respond)
		}
		return result
	}
	return func(w http.ResponseWriter, r *http.Request) {
		query, err := getSearchEntityQueryParams(r.URL.Query())
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

		found, err := logic.Entity.Find(query)
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
		Data *types.EntityRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		entityID, _ := primitive.ObjectIDFromHex(vars["entityID"])
		entity, err := logic.Entity.FindByID(entityID)
		if err != nil {
			l.Logger.Info("[INFO] EntityHandler.getEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}
		api.Respond(w, r, http.StatusOK, respond{Data: types.NewEntityRespondWithoutEmail(entity)})
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
	if !util.IsTradingAccepted(SenderEntity.Status) {
		return errors.New("Sender is not in the accepted status.")

	}
	if !util.IsTradingAccepted(ReceiverEntity.Status) {
		return errors.New("Receiver is not in the accepted status.")
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

		w.WriteHeader(http.StatusOK)
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
