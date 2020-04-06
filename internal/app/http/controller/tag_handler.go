package controller

import (
	"net/http"
	"sync"

	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"

	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/internal/app/api"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/l"
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
		query, err := api.NewSearchTagQuery(r.URL.Query())
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
