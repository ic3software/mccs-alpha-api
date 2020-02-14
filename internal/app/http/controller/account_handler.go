package controller

import (
	"sync"

	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/e"

	"github.com/ic3network/mccs-alpha-api/internal/app/types"
)

type accountHandler struct {
	once *sync.Once
}

var AccountHandler = newAccountHandler()

func newAccountHandler() *accountHandler {
	return &accountHandler{
		once: new(sync.Once),
	}
}

func (a *accountHandler) RegisterRoutes(
	public *mux.Router,
	private *mux.Router,
	adminPublic *mux.Router,
	adminPrivate *mux.Router,
) {
	a.once.Do(func() {
		// adminPrivate.Path("/accounts/search").HandlerFunc(a.searchAccount()).Methods("GET")
	})
}

type searchAccountFormData struct {
	TagType          string
	Tags             []*types.TagField
	CreatedOnOrAfter string
	Status           string
	EntityName       string
	LocationCity     string
	LocationCountry  string
	Category         string
	LastName         string
	Email            string
	Filter           string
	Page             int
}

type account struct {
	Entity  *types.Entity
	User    *types.User
	Balance float64
}

type findAccountResult struct {
	Accounts        []account
	NumberOfResults int
	TotalPages      int
}

type sreachResponse struct {
	FormData  searchAccountFormData
	AdminTags []*types.AdminTag
	Result    *findAccountResult
}

// func (a *accountHandler) searchAccount() func(http.ResponseWriter, *http.Request) {
// 	t := template.NewView("admin/accounts")
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		q := r.URL.Query()

// 		page, err := strconv.Atoi(q.Get("page"))
// 		if err != nil {
// 			l.Logger.Error("SearchAccount failed", zap.Error(err))
// 			t.Error(w, r, nil, err)
// 			return
// 		}

// 		f := searchAccountFormData{
// 			TagType:          q.Get("tag_type"),
// 			Tags:             helper.ToSearchTags(q.Get("tags")),
// 			Status:           q.Get("status"),
// 			EntityName:       q.Get("entity_name"),
// 			CreatedOnOrAfter: q.Get("created_on_or_after"),
// 			LocationCity:     q.Get("location_city"),
// 			LocationCountry:  q.Get("location_country"),
// 			Category:         q.Get("category"),
// 			LastName:         q.Get("last_name"),
// 			Email:            q.Get("email"),
// 			Filter:           q.Get("filter"),
// 			Page:             page,
// 		}
// 		res := sreachResponse{FormData: f, Result: new(findAccountResult)}

// 		if f.Filter != "entity" && f.LastName == "" && f.Email == "" {
// 			t.Render(w, r, res, []string{"Please enter at least one search criteria."})
// 			return
// 		}

// 		adminTags, err := logic.AdminTag.GetAll()
// 		if err != nil {
// 			l.Logger.Error("SearchAccount failed", zap.Error(err))
// 			t.Error(w, r, nil, err)
// 			return
// 		}
// 		res.AdminTags = adminTags

// 		// Search All Status
// 		var status []string
// 		if f.Status == constant.ALL {
// 			status = []string{
// 				constant.Entity.Pending,
// 				constant.Entity.Accepted,
// 				constant.Entity.Rejected,
// 				constant.Trading.Pending,
// 				constant.Trading.Accepted,
// 				constant.Trading.Rejected,
// 			}
// 		} else {
// 			status = []string{f.Status}
// 		}

// 		findResult := new(types.FindEntityResult)
// 		if f.Filter == "entity" {
// 			c := types.SearchCriteria{
// 				TagType:          f.TagType,
// 				Tags:             f.Tags,
// 				Statuses:         status,
// 				EntityName:       f.EntityName,
// 				CreatedOnOrAfter: util.ParseTime(f.CreatedOnOrAfter),
// 				LocationCity:     f.LocationCity,
// 				LocationCountry:  f.LocationCountry,
// 				AdminTag:         f.Category,
// 			}
// 			findResult, err = logic.Entity.FindEntity(&c, int64(f.Page))
// 			if err != nil {
// 				l.Logger.Error("SearchAccount failed", zap.Error(err))
// 				t.Error(w, r, res, err)
// 				return
// 			}
// 			res.Result.TotalPages = findResult.TotalPages
// 			res.Result.NumberOfResults = findResult.NumberOfResults
// 		}

// 		accounts := make([]account, 0)
// 		// Find the user and account balance using entity id.
// 		for _, entity := range findResult.Entities {
// 			user, err := logic.User.FindByEntityID(entity.ID)
// 			if err != nil {
// 				l.Logger.Error("SearchAccount failed", zap.Error(err))
// 				t.Error(w, r, res, err)
// 				return
// 			}
// 			acc, err := logic.Account.FindByEntityID(entity.ID.Hex())
// 			if err != nil {
// 				l.Logger.Error("SearchAccount failed", zap.Error(err))
// 				t.Error(w, r, res, err)
// 				return
// 			}
// 			accounts = append(accounts, account{
// 				Entity:  entity,
// 				User:    user,
// 				Balance: acc.Balance,
// 			})
// 		}
// 		res.Result.Accounts = accounts

// 		if len(res.Result.Accounts) > 0 || f.Filter == "entity" {
// 			t.Render(w, r, res, nil)
// 			return
// 		}

// 		// The logic for searching by user last name and email.
// 		u := types.User{
// 			LastName: f.LastName,
// 			Email:    f.Email,
// 		}
// 		findUserResult, err := logic.User.FindUsers(&u, int64(f.Page))
// 		if err != nil {
// 			l.Logger.Error("SearchAccount failed", zap.Error(err))
// 			t.Error(w, r, res, err)
// 			return
// 		}
// 		res.Result.TotalPages = findUserResult.TotalPages
// 		res.Result.NumberOfResults = findUserResult.NumberOfResults

// 		// Find the entity and account balance.
// 		for _, user := range findUserResult.Users {
// 			entity, err := logic.Entity.FindByID(user.Entities[0])
// 			if err != nil {
// 				l.Logger.Error("SearchAccount failed", zap.Error(err))
// 				t.Error(w, r, res, err)
// 				return
// 			}
// 			acc, err := logic.Account.FindByEntityID(entity.ID.Hex())
// 			if err != nil {
// 				l.Logger.Error("SearchAccount failed", zap.Error(err))
// 				t.Error(w, r, res, err)
// 				return
// 			}
// 			accounts = append(accounts, account{
// 				Entity:  entity,
// 				User:    user,
// 				Balance: acc.Balance,
// 			})
// 		}
// 		res.Result.Accounts = accounts

// 		t.Render(w, r, res, nil)
// 	}
// }

func (a *accountHandler) FindByUserID(uID string) (*types.Account, error) {
	entity, err := EntityHandler.FindByUserID(uID)
	if err != nil {
		return nil, e.Wrap(err, "controller.Entity.FindByUserID failed")
	}
	account, err := logic.Account.FindByEntityID(entity.ID.Hex())
	if err != nil {
		return nil, e.Wrap(err, "controller.Entity.FindByUserID failed")
	}
	return account, nil
}
