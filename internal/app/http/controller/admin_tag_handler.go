package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"github.com/ic3network/mccs-alpha-api/internal/app/api"
	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"

	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/l"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type adminTagHandler struct {
	once *sync.Once
}

var AdminTagHandler = NewAdminTagHandler()

func NewAdminTagHandler() *adminTagHandler {
	return &adminTagHandler{
		once: new(sync.Once),
	}
}

func (handler *adminTagHandler) RegisterRoutes(
	adminPublic *mux.Router,
	adminPrivate *mux.Router,
) {
	handler.once.Do(func() {
		adminPrivate.Path("/tags").HandlerFunc(handler.create()).Methods("POST")
		adminPrivate.Path("/tags/{id}").HandlerFunc(handler.renameTag()).Methods("PATCH")
		adminPrivate.Path("/tags/{id}").HandlerFunc(handler.deleteTag()).Methods("DELETE")
	})
}

func (h *adminTagHandler) create() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.TagRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := types.NewAdminCreateTagReqBody(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		_, err := logic.Tag.FindByName(req.Name)
		if err == nil {
			api.Respond(w, r, http.StatusBadRequest, errors.New("Tag already exists."))
			return
		}

		created, err := logic.Tag.Create(req.Name)
		if err != nil {
			l.Logger.Error("[Error] AdminTagHandler.create failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		api.Respond(w, r, http.StatusOK, respond{Data: &types.TagRespond{
			ID:   created.ID.Hex(),
			Name: created.Name,
		}})
	}
}

func (h *adminTagHandler) renameTag() func(http.ResponseWriter, *http.Request) {
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

func (h *adminTagHandler) deleteTag() func(http.ResponseWriter, *http.Request) {
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
