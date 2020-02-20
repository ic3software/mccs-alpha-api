package controller

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"

	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/api"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/helper"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/l"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/log"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type tagHandler struct {
	once *sync.Once
}

var TagHandler = newTagHandler()

func newTagHandler() *tagHandler {
	return &tagHandler{
		once: new(sync.Once),
	}
}

func (h *tagHandler) RegisterRoutes(
	public *mux.Router,
	private *mux.Router,
	adminPublic *mux.Router,
	adminPrivate *mux.Router,
) {
	h.once.Do(func() {
		public.Path("/api/v1/tags").HandlerFunc(h.searchTag()).Methods("GET")

		adminPrivate.Path("/api/user-tags").HandlerFunc(h.createTag()).Methods("POST")
		adminPrivate.Path("/api/user-tags/{id}").HandlerFunc(h.renameTag()).Methods("PUT")
		adminPrivate.Path("/api/user-tags/{id}").HandlerFunc(h.deleteTag()).Methods("DELETE")
	})
}

func (t *tagHandler) searchTag() func(http.ResponseWriter, *http.Request) {
	type meta struct {
		NumberOfResults int `json:"numberOfResults"`
		TotalPages      int `json:"totalPages"`
	}
	type respond struct {
		Data []string `json:"data"`
		Meta meta     `json:"meta"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		query := &types.SearchTagQuery{
			Fragment: q.Get("fragment"),
			Page:     q.Get("page"),
			PageSize: q.Get("page_size"),
		}

		errs := query.Validate()
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		found, err := logic.Tag.Find(query)
		if err != nil {
			l.Logger.Error("[Error] TagHandler.searchTag failed:", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		api.Respond(w, r, http.StatusOK, respond{
			Data: utils.TagToNames(found.Tags),
			Meta: meta{
				TotalPages:      found.TotalPages,
				NumberOfResults: found.NumberOfResults,
			},
		})
	}
}

// TO BE REMOVED

func (h *tagHandler) SaveOfferTags(added []string) error {
	for _, tagName := range added {
		// TODO: UpdateOffers
		err := logic.Tag.UpdateOffer(tagName)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *tagHandler) SaveWantTags(added []string) error {
	for _, tagName := range added {
		// TODO: UpdateWants
		err := logic.Tag.UpdateWant(tagName)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *tagHandler) createTag() func(http.ResponseWriter, *http.Request) {
	type request struct {
		Name string `json:"name"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			l.Logger.Error("CreateTag failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}

		if req.Name == "" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Please enter the tag name"))
			return
		}

		tagNames := helper.GetTags(req.Name)
		if len(tagNames) == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Please enter a valid tag name"))
			return
		}

		tagName := tagNames[0].Name
		_, err = logic.Tag.FindByName(tagName)
		if err == nil {
			l.Logger.Info("[CreateTag] failed: Tag already exists")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Tag already exists!"))
			return
		}

		err = logic.Tag.Create(tagName)
		if err != nil {
			l.Logger.Error("CreateTag failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}

		go func() {
			objID, _ := primitive.ObjectIDFromHex(r.Header.Get("userID"))
			adminUser, err := logic.AdminUser.FindByID(objID)
			if err != nil {
				l.Logger.Error("log.Admin.CreateTag failed", zap.Error(err))
				return
			}
			err = logic.UserAction.Log(log.Admin.CreateTag(adminUser, tagName))
			if err != nil {
				l.Logger.Error("log.Admin.CreateTag failed", zap.Error(err))
			}
		}()

		w.WriteHeader(http.StatusCreated)
	}
}

func (h *tagHandler) renameTag() func(http.ResponseWriter, *http.Request) {
	type request struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			l.Logger.Error("RenameTag failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}

		if req.Name == "" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Please enter the tag name"))
			return
		}

		_, err = logic.Tag.FindByName(req.Name)
		if err == nil {
			l.Logger.Info("[RenameTag] failed: Tag already exists")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Tag already exists!"))
			return
		}

		tagID, err := primitive.ObjectIDFromHex(req.ID)
		if err != nil {
			l.Logger.Error("RenameTag failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}

		tag, err := logic.Tag.FindByID(tagID)
		if err != nil {
			l.Logger.Error("RenameTag failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Tag not found."))
			return
		}
		oldName := tag.Name

		go func() {
			err := logic.Entity.RenameTag(oldName, req.Name)
			if err != nil {
				l.Logger.Error("RenameTag failed", zap.Error(err))
			}
		}()

		tag = &types.Tag{
			ID:   tagID,
			Name: req.Name,
		}
		err = logic.Tag.Rename(tag)
		if err != nil {
			l.Logger.Error("RenameTag failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}

		go func() {
			objID, _ := primitive.ObjectIDFromHex(r.Header.Get("userID"))
			adminUser, err := logic.AdminUser.FindByID(objID)
			if err != nil {
				l.Logger.Error("log.Admin.ModifyTag failed", zap.Error(err))
				return
			}
			err = logic.UserAction.Log(log.Admin.ModifyTag(adminUser, oldName, req.Name))
			if err != nil {
				l.Logger.Error("log.Admin.ModifyTag failed", zap.Error(err))
			}
		}()

		w.WriteHeader(http.StatusCreated)
	}
}

func (h *tagHandler) deleteTag() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		tagID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			l.Logger.Error("DeleteTag failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tag, err := logic.Tag.FindByID(tagID)
		if err != nil {
			l.Logger.Error("DeleteTag failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = logic.Tag.DeleteByID(tagID)
		if err != nil {
			l.Logger.Error("DeleteTag failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		go func() {
			err := logic.Entity.DeleteTag(tag.Name)
			if err != nil {
				l.Logger.Error("DeleteTag failed", zap.Error(err))
			}
		}()
		go func() {
			objID, _ := primitive.ObjectIDFromHex(r.Header.Get("userID"))
			adminUser, err := logic.AdminUser.FindByID(objID)
			if err != nil {
				l.Logger.Error("log.Admin.DeleteTag failed", zap.Error(err))
				return
			}
			err = logic.UserAction.Log(log.Admin.DeleteTag(adminUser, tag.Name))
			if err != nil {
				l.Logger.Error("log.Admin.DeleteTag failed", zap.Error(err))
			}
		}()

		w.WriteHeader(http.StatusOK)
	}
}
