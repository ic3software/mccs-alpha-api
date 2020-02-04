package logic

import (
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/es"
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/mongo"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type trading struct{}

var Trading = &trading{}

func (t *trading) UpdateEntity(id primitive.ObjectID, data *types.TradingRegisterData) error {
	err := es.Entity.UpdateTradingInfo(id, data)
	if err != nil {
		return err
	}
	err = mongo.Entity.UpdateTradingInfo(id, data)
	if err != nil {
		return err
	}
	return nil
}

func (t *trading) UpdateUser(id primitive.ObjectID, data *types.TradingRegisterData) error {
	err := es.User.UpdateTradingInfo(id, data)
	if err != nil {
		return err
	}
	err = mongo.User.UpdateTradingInfo(id, data)
	if err != nil {
		return err
	}
	return nil
}
