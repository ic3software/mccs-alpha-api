package es

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ic3network/mccs-alpha-api/global/constant"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/helper"
	"github.com/ic3network/mccs-alpha-api/util"
	"github.com/olivere/elastic/v7"
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

func (es *entity) Create(id primitive.ObjectID, data *types.Entity) error {
	body := types.EntityESRecord{
		EntityID:        id.Hex(),
		EntityName:      data.EntityName,
		Offers:          data.Offers,
		Wants:           data.Wants,
		LocationCity:    data.LocationCity,
		LocationCountry: data.LocationCountry,
		Status:          constant.Entity.Pending,
		Categories:      data.Categories,
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

func (es *entity) Update(update *types.Entity) error {
	doc := types.EntityESRecord{
		EntityName:      update.EntityName,
		LocationCountry: update.LocationCity,
		LocationCity:    update.LocationCountry,
	}
	_, err := es.c.Update().
		Index(es.index).
		Id(update.ID.Hex()).
		Doc(doc).
		Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (es *entity) AdminUpdate(update *types.Entity) error {
	doc := types.EntityESRecord{
		EntityName:      update.EntityName,
		LocationCountry: update.LocationCity,
		LocationCity:    update.LocationCountry,
		Status:          update.Status,
	}
	_, err := es.c.Update().
		Index(es.index).
		Id(update.ID.Hex()).
		Doc(doc).
		Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (es *entity) UpdateTags(id primitive.ObjectID, difference *types.TagDifference) error {
	params := map[string]interface{}{
		"offersAdded":   helper.ToTagFields(difference.NewAddedOffers),
		"wantsAdded":    helper.ToTagFields(difference.NewAddedWants),
		"offersRemoved": difference.OffersRemoved,
		"wantsRemoved":  difference.WantsRemoved,
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

func seachByStatus(q *elastic.BoolQuery, req *types.SearchEntityReqBody) *elastic.BoolQuery {
	if len(req.Statuses) != 0 {
		qq := elastic.NewBoolQuery()
		for _, status := range req.Statuses {
			qq.Should(elastic.NewMatchQuery("status", status))
		}
		q.Must(qq)
	}
	return q
}

func seachByAddress(q *elastic.BoolQuery, req *types.SearchEntityReqBody) *elastic.BoolQuery {
	if req.EntityName != "" {
		q.Must(newFuzzyWildcardQuery("entityName", req.EntityName))
	}
	if req.LocationCountry != "" {
		q.Must(elastic.NewMatchQuery("locationCountry", req.LocationCountry))
	}
	if req.LocationCity != "" {
		q.Must(newFuzzyWildcardQuery("locationCity", req.LocationCity))
	}
	return q
}

func seachByTags(q *elastic.BoolQuery, req *types.SearchEntityReqBody) *elastic.BoolQuery {
	// "Tag Added After" will associate with "tags".
	if len(req.Offers) != 0 {
		qq := elastic.NewBoolQuery()
		// weighted is used to make sure the tags are shown in order.
		weighted := 2.0
		for _, offer := range req.Offers {
			qq.Should(newFuzzyWildcardTimeQueryForTag("offers", offer, req.TaggedSince).
				Boost(weighted))
			weighted *= 0.9
		}
		// Must match one of the "Should" queries.
		q.Must(qq)
	}
	if len(req.Wants) != 0 {
		qq := elastic.NewBoolQuery()
		// weighted is used to make sure the tags are shown in order.
		weighted := 2.0
		for _, want := range req.Wants {
			qq.Should(newFuzzyWildcardTimeQueryForTag("wants", want, req.TaggedSince).
				Boost(weighted))
			weighted *= 0.9
		}
		// Must match one of the "Should" queries.
		q.Must(qq)
	}
	return q
}

func (es *entity) Find(req *types.SearchEntityReqBody) (*types.ESFindEntityResult, error) {
	var ids []string

	q := elastic.NewBoolQuery()

	if req.FavoritesOnly {
		idQuery := elastic.NewIdsQuery().Ids(util.ToIDStrings(req.FavoriteEntities)...)
		q.Must(idQuery)
	}
	if req.Category != "" {
		q.Must(elastic.NewMatchQuery("categories", req.Category))
	}

	seachByStatus(q, req)
	seachByAddress(q, req)
	seachByTags(q, req)

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
		ids = append(ids, record.EntityID)
	}

	numberOfResults := int(res.Hits.TotalHits.Value)
	totalPages := util.GetNumberOfPages(numberOfResults, req.PageSize)

	return &types.ESFindEntityResult{
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

// TO BE REMOVED

func (es *entity) UpdateTradingInfo(id primitive.ObjectID, data *types.TradingRegisterData) error {
	doc := map[string]interface{}{
		"entityName":      data.EntityName,
		"locationCity":    data.LocationCity,
		"locationCountry": data.LocationCountry,
		"status":          constant.Trading.Pending,
	}
	_, err := es.c.Update().
		Index(es.index).
		Id(id.Hex()).
		Doc(doc).
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
		return err
	}
	return nil
}
