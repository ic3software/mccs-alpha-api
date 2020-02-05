package controller

import (
	"encoding/json"
	"net/http"
	"sync"

	"strconv"

	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/global/constant"
	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/email"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/helper"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/l"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/template"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/ic3network/mccs-alpha-api/internal/pkg/util"
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
		public.Path("/entities").HandlerFunc(b.searchEntityPage()).Methods("GET")
		public.Path("/entities/search").HandlerFunc(b.searchEntity()).Methods("GET")
		public.Path("/entityPage/{id}").HandlerFunc(b.entityPage()).Methods("GET")
		private.Path("/entities/search/match-tags").HandlerFunc(b.searhMatchTags()).Methods("GET")

		private.Path("/api/entityStatus").HandlerFunc(b.entityStatus()).Methods("GET")
		private.Path("/api/getEntityName").HandlerFunc(b.getEntityName()).Methods("GET")
		private.Path("/api/tradingMemberStatus").HandlerFunc(b.tradingMemberStatus()).Methods("GET")
		private.Path("/api/contactEntity").HandlerFunc(b.contactEntity()).Methods("POST")
	})
}

func (b *entityHandler) FindByID(entityID string) (*types.Entity, error) {
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

func (b *entityHandler) FindByEmail(email string) (*types.Entity, error) {
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

func (b *entityHandler) FindByUserID(uID string) (*types.Entity, error) {
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

type searchEntityFormData struct {
	TagType               string
	Tags                  []*types.TagField
	CreatedOnOrAfter      string
	Category              string
	ShowUserFavoritesOnly bool
	Page                  int
}

type searchEntityResponse struct {
	IsUserLoggedIn   bool
	FormData         searchEntityFormData
	Categories       []string
	Result           *types.FindEntityResult
	FavoriteEntities []primitive.ObjectID
}

func (b *entityHandler) searchEntityPage() func(http.ResponseWriter, *http.Request) {
	t := template.NewView("entities")
	return func(w http.ResponseWriter, r *http.Request) {
		adminTags, err := logic.AdminTag.GetAll()
		if err != nil {
			l.Logger.Error("SearchEntityPage failed", zap.Error(err))
			t.Error(w, r, nil, err)
			return
		}
		res := searchEntityResponse{Categories: helper.GetAdminTagNames(adminTags)}
		_, err = UserHandler.FindByID(r.Header.Get("userID"))
		if err != nil {
			res.IsUserLoggedIn = false
		} else {
			res.IsUserLoggedIn = true
		}
		t.Render(w, r, res, nil)
	}
}

func (b *entityHandler) searchEntity() func(http.ResponseWriter, *http.Request) {
	t := template.NewView("entities")
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		page, err := strconv.Atoi(q.Get("page"))
		if err != nil {
			l.Logger.Error("SearchEntity failed", zap.Error(err))
			t.Error(w, r, nil, err)
			return
		}

		f := searchEntityFormData{
			TagType:               q.Get("tag_type"),
			Tags:                  helper.ToSearchTags(q.Get("tags")),
			CreatedOnOrAfter:      q.Get("created_on_or_after"),
			Category:              q.Get("category"),
			ShowUserFavoritesOnly: q.Get("show-favorites-only") == "true",
			Page:                  page,
		}
		res := searchEntityResponse{FormData: f}

		user, err := UserHandler.FindByID(r.Header.Get("userID"))
		if err != nil {
			res.IsUserLoggedIn = false
		} else {
			res.IsUserLoggedIn = true
			res.FavoriteEntities = user.FavoriteEntities
		}

		c := types.SearchCriteria{
			TagType: f.TagType,
			Tags:    f.Tags,
			Statuses: []string{
				constant.Entity.Accepted,
				constant.Trading.Pending,
				constant.Trading.Accepted,
				constant.Trading.Rejected,
			},
			CreatedOnOrAfter:      util.ParseTime(f.CreatedOnOrAfter),
			AdminTag:              f.Category,
			ShowUserFavoritesOnly: f.ShowUserFavoritesOnly,
			FavoriteEntities:      res.FavoriteEntities,
		}
		findResult, err := logic.Entity.FindEntity(&c, int64(f.Page))
		res.Result = findResult
		if err != nil {
			l.Logger.Error("SearchEntity failed", zap.Error(err))
			t.Error(w, r, res, err)
			return
		}

		adminTags, err := logic.AdminTag.GetAll()
		if err != nil {
			l.Logger.Error("SearchEntity failed", zap.Error(err))
			t.Error(w, r, nil, err)
			return
		}
		res.Categories = helper.GetAdminTagNames(adminTags)

		t.Render(w, r, res, nil)
	}
}

func (b *entityHandler) entityPage() func(http.ResponseWriter, *http.Request) {
	t := template.NewView("entity")
	type formData struct {
		IsUserLoggedIn bool
		EntityEmail    string
		Entity         *types.Entity
		User           *types.User
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bID := vars["id"]
		entity, err := b.FindByID(bID)
		if err != nil {
			l.Logger.Error("EntityPage failed", zap.Error(err))
			t.Error(w, r, nil, err)
			return
		}

		f := formData{
			Entity: entity,
		}

		entityUser, err := UserHandler.FindByEntityID(bID)
		if err != nil {
			l.Logger.Error("EntityPage failed", zap.Error(err))
			t.Error(w, r, nil, err)
			return
		}
		f.EntityEmail = entityUser.Email

		user, err := UserHandler.FindByID(r.Header.Get("userID"))
		if err != nil {
			f.IsUserLoggedIn = false
		} else {
			f.IsUserLoggedIn = true
			f.User = user
		}

		t.Render(w, r, f, nil)
	}
}

func (b *entityHandler) contactEntity() func(http.ResponseWriter, *http.Request) {
	type request struct {
		EntityID string `json:"id"`
		Body     string `json:"body"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			l.Logger.Error("ContactEntity failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}

		user, err := UserHandler.FindByID(r.Header.Get("userID"))
		if err != nil {
			l.Logger.Error("ContactEntity failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}

		entityOwner, err := UserHandler.FindByEntityID(req.EntityID)
		if err != nil {
			l.Logger.Error("ContactEntity failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}

		receiver := entityOwner.FirstName + " " + entityOwner.LastName
		replyToName := user.FirstName + " " + user.LastName
		err = email.SendContactEntity(receiver, entityOwner.Email, replyToName, user.Email, req.Body)
		if err != nil {
			l.Logger.Error("ContactEntity failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (b *entityHandler) searhMatchTags() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/entities/search?"+r.URL.Query().Encode(), http.StatusFound)
	}
}

func (b *entityHandler) entityStatus() func(http.ResponseWriter, *http.Request) {
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

func (b *entityHandler) getEntityName() func(http.ResponseWriter, *http.Request) {
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

func (b *entityHandler) tradingMemberStatus() func(http.ResponseWriter, *http.Request) {
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
