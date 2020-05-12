package controller

import (
	"errors"
	"net/http"
	"sync"

	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"

	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/internal/app/api"
	"github.com/ic3network/mccs-alpha-api/util/l"
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

func (handler *tagHandler) RegisterRoutes(
	public *mux.Router,
	private *mux.Router,
	adminPublic *mux.Router,
	adminPrivate *mux.Router,
) {
	handler.once.Do(func() {
		public.Path("/tags").HandlerFunc(handler.searchTag()).Methods("GET")

		adminPrivate.Path("/tags").HandlerFunc(handler.adminCreate()).Methods("POST")
		adminPrivate.Path("/tags/{id}").HandlerFunc(handler.adminUpdate()).Methods("PATCH")
		adminPrivate.Path("/tags/{id}").HandlerFunc(handler.adminDelete()).Methods("DELETE")
	})
}

func (h *tagHandler) UpdateOffers(added []string) error {
	for _, tagName := range added {
		// TODO: UpdateOffers
		err := logic.Tag.UpdateOffer(tagName)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *tagHandler) UpdateWants(added []string) error {
	for _, tagName := range added {
		// TODO: UpdateWants
		err := logic.Tag.UpdateWant(tagName)
		if err != nil {
			return err
		}
	}
	return nil
}

func (handler *tagHandler) searchTag() func(http.ResponseWriter, *http.Request) {
	type meta struct {
		NumberOfResults int `json:"numberOfResults"`
		TotalPages      int `json:"totalPages"`
	}
	type respond struct {
		Data []string `json:"data"`
		Meta meta     `json:"meta"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		query, err := types.NewSearchTagReq(r.URL.Query())
		if err != nil {
			l.Logger.Info("[Info] TagHandler.searchTag failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		errs := query.Validate()
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		found, err := logic.Tag.Search(query)
		if err != nil {
			l.Logger.Error("[Error] TagHandler.searchTag failed:", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		api.Respond(w, r, http.StatusOK, respond{
			Data: types.TagToNames(found.Tags),
			Meta: meta{
				TotalPages:      found.TotalPages,
				NumberOfResults: found.NumberOfResults,
			},
		})
	}
}

func (h *tagHandler) adminCreate() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.TagRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := types.NewAdminCreateTagReq(r)
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

		go logic.UserAction.AdminCreateTag(r.Header.Get("userID"), req.Name)

		api.Respond(w, r, http.StatusOK, respond{Data: &types.TagRespond{
			ID:   created.ID.Hex(),
			Name: created.Name,
		}})
	}
}

func (h *tagHandler) adminUpdate() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.TagRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := types.NewAdminUpdateTagReq(r)
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

		go logic.UserAction.AdminModifyTag(r.Header.Get("userID"), old.Name, updated.Name)

		api.Respond(w, r, http.StatusOK, respond{Data: &types.TagRespond{
			ID:   updated.ID.Hex(),
			Name: updated.Name,
		}})
	}
}

func (h *tagHandler) adminDelete() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.TagRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := types.NewAdminDeleteTagReq(r)
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

		go logic.UserAction.AdminDeleteTag(r.Header.Get("userID"), deleted.Name)

		api.Respond(w, r, http.StatusOK, respond{Data: &types.TagRespond{
			ID:   deleted.ID.Hex(),
			Name: deleted.Name,
		}})
	}
}
