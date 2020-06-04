package es

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ic3network/mccs-alpha-api/global/constant"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/util"
	"github.com/olivere/elastic/v7"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type entity struct {
	c     *elastic.Client
	index string
}

var Entity = &entity{}

func (es *entity) Register(client *elastic.Client) {
	es.c = client
	es.index = "entities"
}

func (es *entity) Create(id primitive.ObjectID, entity *types.Entity) error {
	balance := 0.0
	maxPosBal := viper.GetFloat64("transaction.maxPosBal")
	maxNegBal := viper.GetFloat64("transaction.maxNegBal")

	body := types.EntityESRecord{
		ID:     id.Hex(),
		Name:   entity.Name,
		Email:  entity.Email,
		Status: constant.Entity.Pending,
		// Tags
		Offers:     entity.Offers,
		Wants:      entity.Wants,
		Categories: entity.Categories,
		// Address
		City:    entity.City,
		Region:  entity.Region,
		Country: entity.Country,
		// Account
		AccountNumber: entity.AccountNumber,
		Balance:       &balance,
		MaxPosBal:     &maxPosBal,
		MaxNegBal:     &maxNegBal,
	}
	_, err := es.c.Index().
		Index(es.index).
		Id(id.Hex()).
		BodyJson(body).
		Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}

// PATCH /user/entities/{entityID}

func (es *entity) Update(req *types.UpdateUserEntityReq) error {
	doc := types.EntityESRecord{
		Name:  req.EntityName,
		Email: req.Email,
		// Address
		City:    req.LocationCity,
		Region:  req.LocationRegion,
		Country: req.LocationCountry,
	}

	script := es.getUpateTagScript(req.AddedOffers, req.AddedWants, req.RemovedOffers, req.RemovedWants)

	// TODO
	// Can not update doc and script together.
	_, err := es.c.Update().
		Index(es.index).
		Id(req.OriginEntity.ID.Hex()).
		Doc(doc).
		Do(context.Background())
	if err != nil {
		return err
	}
	_, err = es.c.Update().
		Index(es.index).
		Id(req.OriginEntity.ID.Hex()).
		Script(script).
		Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}

// PATCH /admin/entities/{entityID}

func (es *entity) AdminUpdate(req *types.AdminUpdateEntityReq) error {
	doc := types.EntityESRecord{
		Name:   req.EntityName,
		Email:  req.Email,
		Status: req.Status,
		// Address
		City:    req.LocationCity,
		Region:  req.LocationRegion,
		Country: req.LocationCountry,
		// Account
		MaxNegBal: req.MaxNegBal,
		MaxPosBal: req.MaxPosBal,
	}

	script := es.getUpateTagScript(req.AddedOffers, req.AddedWants, req.RemovedOffers, req.RemovedWants)

	// TODO
	// Can not update doc and script together.
	_, err := es.c.Update().
		Index(es.index).
		Id(req.OriginEntity.ID.Hex()).
		Doc(doc).
		Do(context.Background())
	if err != nil {
		return err
	}
	_, err = es.c.Update().
		Index(es.index).
		Id(req.OriginEntity.ID.Hex()).
		Script(script).
		Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}

// PATCH /user/entities/{entityID}
// PATCH /admin/entities/{entityID}

func (es *entity) getUpateTagScript(addedOffers []string, addedWants []string, removedOffers []string, removedWants []string) *elastic.Script {
	params := map[string]interface{}{
		"offersAdded":   types.ToTagFields(addedOffers),
		"wantsAdded":    types.ToTagFields(addedWants),
		"offersRemoved": removedOffers,
		"wantsRemoved":  removedWants,
	}

	script := elastic.
		NewScript(`
			if (ctx._source.offers === null) {
				ctx._source["offers"] = [];
			}
			if (ctx._source.wants === null) {
				ctx._source["wants"] = [];
			}

			if (params.offersRemoved !== null && params.offersRemoved.length !== 0) {
				for (int i = 0; i < ctx._source.offers.length; i++) {
					if (params.offersRemoved.contains(ctx._source.offers[i].name)) {
						ctx._source.offers.remove(i);
						i--
					}
				}
			}
			if (params.wantsRemoved !== null && params.wantsRemoved.length !== 0) {
				for (int i = 0; i < ctx._source.wants.length; i++) {
					if (params.wantsRemoved.contains(ctx._source.wants[i].name)) {
						ctx._source.wants.remove(i);
						i--
					}
				}
			}

			if (params.offersAdded !== null && params.offersAdded.length !== 0) {
				for (int i = 0; i < params.offersAdded.length; i++) {
					ctx._source.offers.add(params.offersAdded[i]);
				}
			}
			if (params.wantsAdded !== null && params.wantsAdded.length !== 0) {
				for (int i = 0; i < params.wantsAdded.length; i++) {
					ctx._source.wants.add(params.wantsAdded[i]);
				}
			}
		`).Params(params)

	return script
}

// GET /entities

func (es *entity) Search(req *types.SearchEntityReq) (*types.ESSearchEntityResult, error) {
	var ids []string

	q := elastic.NewBoolQuery()

	if req.FavoritesOnly {
		idQuery := elastic.NewIdsQuery().Ids(util.ToIDStrings(req.FavoriteEntities)...)
		q.Must(idQuery)
	}
	if req.Category != "" {
		q.Must(elastic.NewMatchQuery("categories", req.Category))
	}

	seachByStatus(q, req.Statuses)
	seachbyNameEmailAndAddress(q, &byNameAndAddress{
		Name:    req.EntityName,
		City:    req.LocationCity,
		Country: req.LocationCountry,
	})
	seachByTags(q, &byTag{
		Offers:      req.Offers,
		Wants:       req.Wants,
		TaggedSince: req.TaggedSince,
	})

	from := req.PageSize * (req.Page - 1)
	res, err := es.c.Search().
		Index(es.index).
		From(from).
		Size(req.PageSize).
		Query(q).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	for _, hit := range res.Hits.Hits {
		var record types.EntityESRecord
		err := json.Unmarshal(hit.Source, &record)
		if err != nil {
			return nil, err
		}
		ids = append(ids, record.ID)
	}

	numberOfResults := int(res.Hits.TotalHits.Value)
	totalPages := util.GetNumberOfPages(numberOfResults, req.PageSize)

	return &types.ESSearchEntityResult{
		IDs:             ids,
		NumberOfResults: int(numberOfResults),
		TotalPages:      totalPages,
	}, nil
}

func seachByStatus(q *elastic.BoolQuery, status []string) *elastic.BoolQuery {
	if len(status) != 0 {
		qq := elastic.NewBoolQuery()
		for _, status := range status {
			qq.Should(elastic.NewMatchQuery("status", status))
		}
		q.Must(qq)
	}
	return q
}

type byNameAndAddress struct {
	Name    string
	Email   string
	City    string
	Region  string
	Country string
}

func seachbyNameEmailAndAddress(q *elastic.BoolQuery, req *byNameAndAddress) {
	if req.Name != "" {
		q.Must(newFuzzyWildcardQuery("name", req.Name))
	}
	if req.Email != "" {
		q.Must(newWildcardQuery("email", req.Email))
	}
	if req.City != "" {
		q.Must(newFuzzyWildcardQuery("city", req.City))
	}
	if req.Region != "" {
		q.Must(newFuzzyWildcardQuery("region", req.Region))
	}
	if req.Country != "" {
		q.Must(elastic.NewMatchQuery("country", req.Country))
	}
}

type byTag struct {
	Offers      []string
	Wants       []string
	TaggedSince time.Time
}

func seachByTags(q *elastic.BoolQuery, query *byTag) {
	// "Tag Added After" will associate with "tags".
	if len(query.Offers) != 0 {
		qq := elastic.NewBoolQuery()
		// weighted is used to make sure the tags are shown in order.
		weighted := 2.0
		for _, offer := range query.Offers {
			qq.Should(newFuzzyWildcardTimeQueryForTag("offers", offer, query.TaggedSince).
				Boost(weighted))
			weighted *= 0.9
		}
		// Must match one of the "Should" queries.
		q.Must(qq)
	}
	if len(query.Wants) != 0 {
		qq := elastic.NewBoolQuery()
		// weighted is used to make sure the tags are shown in order.
		weighted := 2.0
		for _, want := range query.Wants {
			qq.Should(newFuzzyWildcardTimeQueryForTag("wants", want, query.TaggedSince).
				Boost(weighted))
			weighted *= 0.9
		}
		// Must match one of the "Should" queries.
		q.Must(qq)
	}
}

// GET /admin/entities

func seachByAccount(q *elastic.BoolQuery, req *byAccount) {
	if req.AccountNumber != "" {
		q.Must(elastic.NewMatchQuery("accountNumber", req.AccountNumber))
	}
	if req.Balance != nil {
		q.Must(elastic.NewRangeQuery("balance").Gte(*req.Balance).Lte(*req.Balance))
	}
	if req.MaxPosBal != nil {
		q.Must(elastic.NewRangeQuery("maxPosBal").Gte(*req.MaxPosBal))
	}

	if req.MaxNegBal != nil {
		q.Must(elastic.NewRangeQuery("maxNegBal").Gte(*req.MaxNegBal))
	}
}

type byAccount struct {
	AccountNumber string
	Balance       *float64
	MaxNegBal     *float64
	MaxPosBal     *float64
}

func (es *entity) AdminSearch(req *types.AdminSearchEntityReq) (*types.ESSearchEntityResult, error) {
	var ids []string

	q := elastic.NewBoolQuery()

	if req.Category != "" {
		q.Must(elastic.NewMatchQuery("categories", req.Category))
	}

	seachByStatus(q, req.Statuses)
	seachbyNameEmailAndAddress(q, &byNameAndAddress{
		Name:    req.EntityName,
		Email:   req.EntityEmail,
		City:    req.City,
		Region:  req.Region,
		Country: req.Country,
	})
	seachByTags(q, &byTag{
		Offers:      req.Offers,
		Wants:       req.Wants,
		TaggedSince: req.TaggedSince,
	})
	seachByAccount(q, &byAccount{
		AccountNumber: req.AccountNumber,
		Balance:       req.Balance,
		MaxPosBal:     req.MaxPosBal,
		MaxNegBal:     req.MaxNegBal,
	})

	from := req.PageSize * (req.Page - 1)
	res, err := es.c.Search().
		Index(es.index).
		From(from).
		Size(req.PageSize).
		Query(q).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	for _, hit := range res.Hits.Hits {
		var record types.EntityESRecord
		err := json.Unmarshal(hit.Source, &record)
		if err != nil {
			return nil, err
		}
		ids = append(ids, record.ID)
	}

	numberOfResults := int(res.Hits.TotalHits.Value)
	totalPages := util.GetNumberOfPages(numberOfResults, req.PageSize)

	return &types.ESSearchEntityResult{
		IDs:             ids,
		NumberOfResults: int(numberOfResults),
		TotalPages:      totalPages,
	}, nil
}

func (es *entity) UpdateAllTagsCreatedAt(id primitive.ObjectID, t time.Time) error {
	params := map[string]interface{}{
		"createdAt": t,
	}

	script := elastic.
		NewScript(`
			if (ctx._source.offers !== null) {
				for (int i = 0; i < ctx._source.offers.length; i++) {
					ctx._source.offers[i].createdAt = params.createdAt
				}
			}
			if (ctx._source.wants !== null) {
				for (int i = 0; i < ctx._source.wants.length; i++) {
					ctx._source.wants[i].createdAt = params.createdAt
				}
			}
		`).
		Params(params)

	_, err := es.c.Update().
		Index(es.index).
		Id(id.Hex()).
		Script(script).
		Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (es *entity) RenameCategory(old string, new string) error {
	query := elastic.NewMatchQuery("categories", old)
	script := elastic.
		NewScript(`
			if (ctx._source.categories.contains(params.old)) {
				ctx._source.categories.remove(ctx._source.categories.indexOf(params.old));
				ctx._source.categories.add(params.new);
			}
		`).
		Params(map[string]interface{}{"new": new, "old": old})
	_, err := es.c.UpdateByQuery(es.index).
		Query(query).
		Script(script).
		Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (es *entity) DeleteCategory(name string) error {
	query := elastic.NewMatchQuery("categories", name)
	script := elastic.
		NewScript(`
			if (ctx._source.categories.contains(params.name)) {
				ctx._source.categories.remove(ctx._source.categories.indexOf(params.name));
			}
		`).
		Params(map[string]interface{}{"name": name})
	_, err := es.c.UpdateByQuery(es.index).
		Query(query).
		Script(script).
		Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (es *entity) RenameTag(old string, new string) error {
	query := elastic.NewBoolQuery()
	query.Should(elastic.NewMatchQuery("offers.name", old))
	query.Should(elastic.NewMatchQuery("wants.name", old))
	script := elastic.
		NewScript(`
			for (int i = 0; i < ctx._source.offers.length; i++) {
				if (ctx._source.offers[i].name == params.old) {
					ctx._source.offers[i].name = params.new
				}
			}
			for (int i = 0; i < ctx._source.wants.length; i++) {
				if (ctx._source.wants[i].name == params.old) {
					ctx._source.wants[i].name = params.new
				}
			}
		`).
		Params(map[string]interface{}{"new": new, "old": old})
	_, err := es.c.UpdateByQuery(es.index).
		Query(query).
		Script(script).
		Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (es *entity) DeleteTag(name string) error {
	query := elastic.NewBoolQuery()
	query.Should(elastic.NewMatchQuery("offers.name", name))
	query.Should(elastic.NewMatchQuery("wants.name", name))
	script := elastic.
		NewScript(`
			for (int i = 0; i < ctx._source.offers.length; i++) {
				if (ctx._source.offers[i].name == params.name) {
					ctx._source.offers.remove(i);
					break;
				}
			}
			for (int i = 0; i < ctx._source.wants.length; i++) {
				if (ctx._source.wants[i].name == params.name) {
					ctx._source.wants.remove(i);
					break;
				}
			}
		`).
		Params(map[string]interface{}{"name": name})
	_, err := es.c.UpdateByQuery(es.index).
		Query(query).
		Script(script).
		Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (es *entity) Delete(id string) error {
	_, err := es.c.Delete().
		Index(es.index).
		Id(id).
		Do(context.Background())
	if err != nil {
		if elastic.IsNotFound(err) {
			return errors.New("Entity does not exist.")
		}
		return err
	}
	return nil
}

// PATCH /transfers/{transferID}

func (es *entity) UpdateBalance(accountNumber string, balance float64) error {
	query := elastic.NewMatchQuery("accountNumber", accountNumber)
	script := elastic.
		NewScript(`ctx._source.balance= params.balance`).
		Params(map[string]interface{}{"balance": balance})
	_, err := es.c.UpdateByQuery(es.index).
		Query(query).
		Script(script).
		Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}
