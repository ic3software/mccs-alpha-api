package controller

import (
	"errors"
	"net/http"

	"sync"

	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/internal/app/api"
	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/util/l"
	"go.uber.org/zap"
)

type categoryHandler struct {
	once *sync.Once
}

var CategoryHandler = newCategoryHandler()

func newCategoryHandler() *categoryHandler {
	return &categoryHandler{
		once: new(sync.Once),
	}
}

func (handler *categoryHandler) RegisterRoutes(
	public *mux.Router,
	private *mux.Router,
	adminPublic *mux.Router,
	adminPrivate *mux.Router,
) {
	handler.once.Do(func() {
		public.Path("/categories").HandlerFunc(handler.search()).Methods("GET")

		adminPrivate.Path("/categories").HandlerFunc(handler.create()).Methods("POST")
		adminPrivate.Path("/categories/{id}").HandlerFunc(handler.update()).Methods("PATCH")
		adminPrivate.Path("/categories/{id}").HandlerFunc(handler.delete()).Methods("DELETE")
	})
}

func (handler *categoryHandler) Update(categories []string) {
	err := logic.Category.Create(categories...)
	if err != nil {
		l.Logger.Error("[Error] CategoryHandler.Update failed:", zap.Error(err))
	}
}

// GET /categories

func (handler *categoryHandler) search() func(http.ResponseWriter, *http.Request) {
	type meta struct {
		NumberOfResults int `json:"numberOfResults"`
		TotalPages      int `json:"totalPages"`
	}
	type respond struct {
		Data []string `json:"data"`
		Meta meta     `json:"meta"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := types.NewSearchCategoryReq(r.URL.Query())
		if err != nil {
			l.Logger.Info("[Info] CategoryHandler.search failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		errs := req.Validate()
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		found, err := logic.Category.Search(req)
		if err != nil {
			l.Logger.Error("[Error] CategoryHandler.search failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		api.Respond(w, r, http.StatusOK, respond{
			Data: handler.categoryToStrings(found.Categories),
			Meta: meta{
				TotalPages:      found.TotalPages,
				NumberOfResults: found.NumberOfResults,
			},
		})
	}
}

func (handler *categoryHandler) categoryToStrings(categories []*types.Category) []string {
	names := []string{}
	for _, c := range categories {
		names = append(names, c.Name)
	}
	return names
}

// POST /admin/categories

func (handler *categoryHandler) create() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.AdminCategoryRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := types.NewAdminCreateCategoryReq(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		_, err := logic.Category.FindByName(req.Name)
		if err == nil {
			api.Respond(w, r, http.StatusBadRequest, errors.New("Category already exists."))
			return
		}

		created, err := logic.Category.CreateOne(req.Name)
		if err != nil {
			l.Logger.Error("[Error] CategoryHandler.create failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		api.Respond(w, r, http.StatusOK, respond{Data: types.NewAdminCategoryRespond(created)})
	}
}

// PATCH /admin/categories/{id}

func (handler *categoryHandler) update() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.AdminCategoryRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := types.NewAdminUpdateCategoryReq(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		_, err := logic.Category.FindByName(req.Name)
		if err == nil {
			api.Respond(w, r, http.StatusBadRequest, errors.New("Category already exists."))
			return
		}

		old, err := logic.Category.FindByIDString(req.ID)
		if err != nil {
			l.Logger.Error("[Error] CategoryHandler.update failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		updated, err := logic.Category.FindOneAndUpdate(old.ID, &types.Category{Name: req.Name})
		if err != nil {
			l.Logger.Error("[Error] CategoryHandler.update failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		go func() {
			err := logic.Entity.RenameCategory(old.Name, updated.Name)
			if err != nil {
				l.Logger.Error("[Error] CategoryHandler.update failed:", zap.Error(err))
				return
			}
		}()

		api.Respond(w, r, http.StatusOK, respond{Data: types.NewAdminCategoryRespond(updated)})
	}
}

// DELETE /admin/categories/{id}

func (handler *categoryHandler) delete() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.AdminCategoryRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := types.NewAdminDeleteCategoryReq(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		deleted, err := logic.Category.FindOneAndDelete(req.ID)
		if err != nil {
			l.Logger.Error("[Error] CategoryHandler.delete failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		go func() {
			err := logic.Entity.DeleteCategory(deleted.Name)
			if err != nil {
				l.Logger.Error("[Error] logic.Entity.delete failed:", zap.Error(err))
			}
		}()

		api.Respond(w, r, http.StatusOK, respond{Data: types.NewAdminCategoryRespond(deleted)})
	}
}
