package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"sync"

	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/internal/app/api"
	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/helper"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/l"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/log"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/template"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (a *categoryHandler) RegisterRoutes(
	public *mux.Router,
	private *mux.Router,
	adminPublic *mux.Router,
	adminPrivate *mux.Router,
) {
	a.once.Do(func() {
		public.Path("/categories").HandlerFunc(a.searchCategory()).Methods("GET")

		adminPrivate.Path("​/categories").HandlerFunc(a.searchAdminTags()).Methods("GET")
		adminPrivate.Path("/api/admin-tags").HandlerFunc(a.createAdminTag()).Methods("POST")
		adminPrivate.Path("/api/admin-tags/{id}").HandlerFunc(a.renameAdminTag()).Methods("PUT")
		adminPrivate.Path("/api/admin-tags/{id}").HandlerFunc(a.deleteAdminTag()).Methods("DELETE")
	})
}

func (handler *categoryHandler) Update(categories []string) {
	err := logic.Category.Create(categories...)
	if err != nil {
		l.Logger.Error("[Error] CategoryHandler.Update failed:", zap.Error(err))
	}
}

func (a *categoryHandler) searchCategory() func(http.ResponseWriter, *http.Request) {
	type meta struct {
		NumberOfResults int `json:"numberOfResults"`
		TotalPages      int `json:"totalPages"`
	}
	type respond struct {
		Data []string `json:"data"`
		Meta meta     `json:"meta"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		query, err := api.NewSearchCategoryQuery(r.URL.Query())
		if err != nil {
			l.Logger.Info("[Info] CategoryHandler.searchCategory failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		errs := query.Validate()
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		found, err := logic.Category.Find(query)
		if err != nil {
			l.Logger.Error("[Error] CategoryHandler.searchCategory failed:", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		api.Respond(w, r, http.StatusOK, respond{
			Data: utils.CategoryToNames(found.Categories),
			Meta: meta{
				TotalPages:      found.TotalPages,
				NumberOfResults: found.NumberOfResults,
			},
		})
	}
}

// TO BE REMOVED

func (a *categoryHandler) SaveAdminTags(adminTags []string) error {
	for _, adminTag := range adminTags {
		err := logic.Category.Create(adminTag)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *categoryHandler) createAdminTag() func(http.ResponseWriter, *http.Request) {
	type request struct {
		Name string `json:"name"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			l.Logger.Error("CreateAdminTag failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}

		if req.Name == "" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Please enter the admin tag name"))
			return
		}
		req.Name = helper.FormatAdminTag(req.Name)

		_, err = logic.Category.FindByName(req.Name)
		if err == nil {
			l.Logger.Info("Admin tag already exists!")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Admin tag already exists!"))
			return
		}

		err = logic.Category.Create(req.Name)
		if err != nil {
			l.Logger.Error("CreateAdminTag failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}

		go func() {
			objID, _ := primitive.ObjectIDFromHex(r.Header.Get("userID"))
			adminUser, err := logic.AdminUser.FindByID(objID)
			if err != nil {
				l.Logger.Error("log.Admin.CreateAdminTag failed", zap.Error(err))
				return
			}
			err = logic.UserAction.Log(log.Admin.CreateAdminTag(adminUser, req.Name))
			if err != nil {
				l.Logger.Error("log.Admin.CreateAdminTag failed", zap.Error(err))
			}
		}()

		w.WriteHeader(http.StatusCreated)
	}
}

func (a *categoryHandler) searchAdminTags() func(http.ResponseWriter, *http.Request) {
	t := template.NewView("admin/admin-tags")
	type formData struct {
		Name string
		Page int
	}
	type response struct {
		FormData formData
		Result   *types.FindCategoryResult
	}
	return func(w http.ResponseWriter, r *http.Request) {
		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			l.Logger.Error("SearchAdminTags failed", zap.Error(err))
			t.Error(w, r, nil, err)
			return
		}

		f := formData{
			Name: r.URL.Query().Get("name"),
			Page: page,
		}
		res := response{FormData: f}

		if f.Name == "" {
			t.Render(w, r, res, []string{"Please enter the admin tag name"})
			return
		}

		findResult, err := logic.Category.FindTags(f.Name, int64(f.Page))
		if err != nil {
			l.Logger.Error("SearchAdminTags failed", zap.Error(err))
			t.Error(w, r, res, err)
			return
		}
		res.Result = findResult

		t.Render(w, r, res, nil)
	}
}

func (a *categoryHandler) renameAdminTag() func(http.ResponseWriter, *http.Request) {
	type request struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			l.Logger.Error("RenameAdminTag failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}

		if req.Name == "" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Please enter the tag name"))
			return
		}
		req.Name = helper.FormatAdminTag(req.Name)

		_, err = logic.Category.FindByName(req.Name)
		if err == nil {
			l.Logger.Info("RenameAdminTag failed: Admin tag already exists")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Admin tag already exists!"))
			return
		}

		adminTagID, err := primitive.ObjectIDFromHex(req.ID)
		if err != nil {
			l.Logger.Error("RenameAdminTag failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}

		adminTag, err := logic.Category.FindByID(adminTagID)
		if err != nil {
			l.Logger.Error("RenameAdminTag failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("AdminTag not found."))
			return
		}
		oldName := adminTag.Name

		go func() {
			err := logic.Entity.RenameAdminTag(oldName, req.Name)
			if err != nil {
				l.Logger.Error("RenameAdminTag failed", zap.Error(err))
			}
		}()

		adminTag = &types.Category{
			ID:   adminTagID,
			Name: req.Name,
		}
		err = logic.Category.Update(adminTag)
		if err != nil {
			l.Logger.Error("RenameAdminTag failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong. Please try again later."))
			return
		}

		go func() {
			objID, _ := primitive.ObjectIDFromHex(r.Header.Get("userID"))
			adminUser, err := logic.AdminUser.FindByID(objID)
			if err != nil {
				l.Logger.Error("log.Admin.ModifyAdminTag failed", zap.Error(err))
				return
			}
			err = logic.UserAction.Log(log.Admin.ModifyAdminTag(adminUser, oldName, req.Name))
			if err != nil {
				l.Logger.Error("log.Admin.ModifyAdminTag failed", zap.Error(err))
			}
		}()

		w.WriteHeader(http.StatusCreated)
	}
}

func (a *categoryHandler) deleteAdminTag() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		adminTagID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			l.Logger.Error("DeleteAdminTag failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		adminTag, err := logic.Category.FindByID(adminTagID)
		if err != nil {
			l.Logger.Error("DeleteAdminTag failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = logic.Category.DeleteByID(adminTagID)
		if err != nil {
			l.Logger.Error("DeleteAdminTag failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		go func() {
			err := logic.Entity.DeleteAdminTags(adminTag.Name)
			if err != nil {
				l.Logger.Error("DeleteAdminTags failed", zap.Error(err))
			}
		}()
		go func() {
			objID, _ := primitive.ObjectIDFromHex(r.Header.Get("userID"))
			adminUser, err := logic.AdminUser.FindByID(objID)
			if err != nil {
				l.Logger.Error("log.Admin.DeleteAdminTag failed", zap.Error(err))
				return
			}
			err = logic.UserAction.Log(log.Admin.DeleteAdminTag(adminUser, adminTag.Name))
			if err != nil {
				l.Logger.Error("log.Admin.DeleteAdminTag failed", zap.Error(err))
			}
		}()

		w.WriteHeader(http.StatusOK)
	}
}
