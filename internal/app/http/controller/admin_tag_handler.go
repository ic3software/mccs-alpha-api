package controller

import (
	"errors"
	"net/http"
	"sync"

	"github.com/ic3network/mccs-alpha-api/internal/app/api"
	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"

	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/l"
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
		adminPrivate.Path("/tags/{id}").HandlerFunc(handler.update()).Methods("PATCH")
		adminPrivate.Path("/tags/{id}").HandlerFunc(handler.delete()).Methods("DELETE")
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

func (h *adminTagHandler) update() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.TagRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := types.NewAdminUpdateTagReqBody(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		_, err := logic.Tag.FindByName(req.Name)
		if err == nil {
			api.Respond(w, r, http.StatusBadRequest, errors.New("Tag already exists."))
			return
		}

		old, err := logic.Tag.FindByIDString(req.ID)
		if err != nil {
			l.Logger.Error("[Error] AdminTagHandler.update failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		updated, err := logic.Tag.FindOneAndUpdate(old.ID, &types.Tag{Name: req.Name})
		if err != nil {
			l.Logger.Error("[Error] AdminTagHandler.update failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		go func() {
			err := logic.Entity.RenameTag(old.Name, updated.Name)
			if err != nil {
				l.Logger.Error("[Error] AdminTagHandler.update failed:", zap.Error(err))
				return
			}
		}()

		api.Respond(w, r, http.StatusOK, respond{Data: &types.TagRespond{
			ID:   updated.ID.Hex(),
			Name: updated.Name,
		}})
	}
}

func (h *adminTagHandler) delete() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.TagRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := types.NewAdminDeleteTagReqBody(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		deleted, err := logic.Tag.FindOneAndDelete(req.ID)
		if err != nil {
			l.Logger.Error("[Error] CategoryHandler.delete failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		go func() {
			err := logic.Entity.DeleteTag(deleted.Name)
			if err != nil {
				l.Logger.Error("[Error] logic.Entity.delete failed:", zap.Error(err))
			}
		}()

		api.Respond(w, r, http.StatusOK, respond{Data: &types.TagRespond{
			ID:   deleted.ID.Hex(),
			Name: deleted.Name,
		}})
	}
}
