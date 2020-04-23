package es

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/util"
	"github.com/olivere/elastic/v7"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type user struct {
	c     *elastic.Client
	index string
}

var User = &user{}

func (es *user) Register(client *elastic.Client) {
	es.c = client
	es.index = "users"
}

func (es *user) Create(userID primitive.ObjectID, user *types.User) error {
	body := types.UserESRecord{
		UserID:    userID.Hex(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}
	_, err := es.c.Index().
		Index(es.index).
		Id(userID.Hex()).
		BodyJson(body).
		Do(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (es *user) Update(userID primitive.ObjectID, update *types.User) error {
	doc := map[string]interface{}{
		"email":     update.Email,
		"firstName": update.FirstName,
		"lastName":  update.LastName,
	}

	_, err := es.c.Update().
		Index(es.index).
		Id(userID.Hex()).
		Doc(doc).
		Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (es *user) Delete(id string) error {
	_, err := es.c.Delete().
		Index(es.index).
		Id(id).
		Do(context.Background())
	if err != nil {
		if elastic.IsNotFound(err) {
			return errors.New("User does not exist.")
		}
	}
	return nil
}

func (es *user) AdminSearchUser(req *types.AdminSearchUserReqBody) (*types.ESSearchUserResult, error) {
	var ids []string
	pageSize := req.PageSize
	from := pageSize * (req.Page - 1)

	q := elastic.NewBoolQuery()

	if req.LastName != "" {
		q.Must(newFuzzyWildcardQuery("lastName", req.LastName))
	}
	if req.Email != "" {
		q.Must(newFuzzyWildcardQuery("email", req.Email))
	}

	res, err := es.c.Search().
		Index(es.index).
		From(from).
		Size(pageSize).
		Query(q).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	for _, hit := range res.Hits.Hits {
		var record types.UserESRecord
		err := json.Unmarshal(hit.Source, &record)
		if err != nil {
			return nil, err
		}
		ids = append(ids, record.UserID)
	}

	numberOfResults := int(res.Hits.TotalHits.Value)
	totalPages := util.GetNumberOfPages(numberOfResults, pageSize)

	return &types.ESSearchUserResult{
		UserIDs:         ids,
		NumberOfResults: numberOfResults,
		TotalPages:      totalPages,
	}, nil
}

// // TO BE REMOVED

// func (es *user) UpdateTradingInfo(id primitive.ObjectID, data *types.TradingRegisterData) error {
// 	doc := map[string]interface{}{
// 		"firstName": data.FirstName,
// 		"lastName":  data.LastName,
// 	}
// 	_, err := es.c.Update().
// 		Index(es.index).
// 		Id(id.Hex()).
// 		Doc(doc).
// 		Do(context.Background())
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
