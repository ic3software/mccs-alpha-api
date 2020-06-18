package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ic3network/mccs-alpha-api/global"
	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/es"
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/mongo"
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/pg"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/util/l"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

func main() {
	global.Init()
	restore()
}

func restore() {
	funcs := []func(){
		restoreEntities,
		restoreUsers,
		restoreTags,
		restoreJournals,
		restoreUserActions,
	}
	for _, f := range funcs {
		f()
	}
}

func restoreEntities() {
	l.Logger.Info("Restoring entities")
	startTime := time.Now()

	cur, err := mongo.DB().Collection("entities").Find(context.TODO(), bson.M{"deletedAt": bson.M{"$exists": false}})
	if err != nil {
		l.Logger.Fatal("[ERROR] restoring entities failed:", zap.Error(err))
		return
	}

	counter := 0
	for cur.Next(context.TODO()) {
		var entity types.Entity
		err := cur.Decode(&entity)
		if err != nil {
			l.Logger.Fatal("[ERROR] restoring entities failed:", zap.Error(err))
			return
		}

		account, err := logic.Account.FindByAccountNumber(entity.AccountNumber)
		if err != nil {
			l.Logger.Fatal("[ERROR] restoring entities failed:", zap.Error(err))
			return
		}
		limit, err := logic.BalanceLimit.FindByAccountNumber(entity.AccountNumber)
		if err != nil {
			l.Logger.Fatal("[ERROR] restoring entities failed:", zap.Error(err))
			return
		}

		// Add the entity to elastic search.
		uRecord := types.EntityESRecord{
			ID:     entity.ID.Hex(),
			Name:   entity.Name,
			Email:  entity.Email,
			Status: entity.Status,
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
			Balance:       &account.Balance,
			MaxNegBal:     &limit.MaxNegBal,
			MaxPosBal:     &limit.MaxPosBal,
		}
		_, err = es.Client().Index().
			Index("entities").
			Id(entity.ID.Hex()).
			BodyJson(uRecord).
			Do(context.Background())
		if err != nil {
			l.Logger.Fatal("[ERROR] restoring entities failed:", zap.Error(err))
			return
		}
		counter++
	}
	if err := cur.Err(); err != nil {
		l.Logger.Fatal("[ERROR] restoring entities failed:", zap.Error(err))
		return
	}
	cur.Close(context.TODO())

	l.Logger.Info(fmt.Sprintf("count %v", counter))
	l.Logger.Info(fmt.Sprintf("took  %v\n\n", time.Now().Sub(startTime)))
}

func restoreUsers() {
	l.Logger.Info("Restoring users")
	startTime := time.Now()

	cur, err := mongo.DB().Collection("users").Find(context.TODO(), bson.M{"deletedAt": bson.M{"$exists": false}})
	if err != nil {
		l.Logger.Fatal("[ERROR] restoring users failed:", zap.Error(err))
		return
	}

	counter := 0
	for cur.Next(context.TODO()) {
		var user types.User
		err := cur.Decode(&user)
		if err != nil {
			l.Logger.Fatal("[ERROR] restoring users failed:", zap.Error(err))
			return
		}
		uRecord := types.UserESRecord{
			UserID:    user.ID.Hex(),
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
		}
		_, err = es.Client().Index().
			Index("users").
			Id(user.ID.Hex()).
			BodyJson(uRecord).
			Do(context.Background())
		if err != nil {
			l.Logger.Fatal("[ERROR] restoring users failed:", zap.Error(err))
			return
		}
		counter++
	}
	if err := cur.Err(); err != nil {
		l.Logger.Fatal("[ERROR] restoring users failed:", zap.Error(err))
		return
	}
	cur.Close(context.TODO())

	l.Logger.Info(fmt.Sprintf("count %v", counter))
	l.Logger.Info(fmt.Sprintf("took  %v\n\n", time.Now().Sub(startTime)))
}

func restoreTags() {
	l.Logger.Info("Restoring tags")
	startTime := time.Now()

	cur, err := mongo.DB().Collection("tags").Find(context.TODO(), bson.M{"deletedAt": bson.M{"$exists": false}})
	if err != nil {
		l.Logger.Fatal("[ERROR] restoring tags failed:", zap.Error(err))
		return
	}
	counter := 0
	for cur.Next(context.TODO()) {
		var tag types.Tag
		err := cur.Decode(&tag)
		if err != nil {
			l.Logger.Fatal("[ERROR] restoring tags failed:", zap.Error(err))
			return
		}
		uRecord := types.TagESRecord{
			TagID:        tag.ID.Hex(),
			Name:         tag.Name,
			OfferAddedAt: tag.OfferAddedAt,
			WantAddedAt:  tag.WantAddedAt,
		}
		_, err = es.Client().Index().
			Index("tags").
			Id(tag.ID.Hex()).
			BodyJson(uRecord).
			Do(context.Background())
		if err != nil {
			l.Logger.Fatal("[ERROR] restoring tags failed:", zap.Error(err))
			return
		}
		counter++
	}
	if err := cur.Err(); err != nil {
		l.Logger.Fatal("[ERROR] restoring tags failed:", zap.Error(err))
		return
	}
	cur.Close(context.TODO())

	l.Logger.Info(fmt.Sprintf("count %v", counter))
	l.Logger.Info(fmt.Sprintf("took  %v\n\n", time.Now().Sub(startTime)))
}

func restoreJournals() {
	l.Logger.Info("Restoring journals")
	startTime := time.Now()

	var journals []*types.Journal
	err := pg.DB().Raw(`
		SELECT *
		FROM journals
		WHERE deleted_at IS NULL
	`).Scan(&journals).Error
	if err != nil {
		l.Logger.Fatal("[ERROR] restoring journals failed:", zap.Error(err))
		return
	}

	for _, journal := range journals {
		record := types.JournalESRecord{
			TransferID:        journal.TransferID,
			FromAccountNumber: journal.FromAccountNumber,
			ToAccountNumber:   journal.ToAccountNumber,
			Status:            journal.Status,
			CreatedAt:         journal.CreatedAt,
		}
		_, err = es.Client().Index().
			Index("journals").
			Id(journal.TransferID).
			BodyJson(record).
			Do(context.Background())
		if err != nil {
			l.Logger.Fatal("[ERROR] restoring journals failed:", zap.Error(err))
			return
		}
	}

	l.Logger.Info(fmt.Sprintf("count %v", len(journals)))
	l.Logger.Info(fmt.Sprintf("took  %v\n\n", time.Now().Sub(startTime)))
}

func restoreUserActions() {
	l.Logger.Info("Restoring user_actions")
	startTime := time.Now()

	cur, err := mongo.DB().Collection("userActions").Find(context.TODO(), bson.M{"deletedAt": bson.M{"$exists": false}})
	if err != nil {
		l.Logger.Fatal("[ERROR] restoring user_actions failed:", zap.Error(err))
		return
	}
	counter := 0
	for cur.Next(context.TODO()) {
		var userAction types.UserAction
		err := cur.Decode(&userAction)
		if err != nil {
			l.Logger.Fatal("[ERROR] restoring user_actions failed:", zap.Error(err))
			return
		}
		record := types.UserActionESRecord{
			UserID:    userAction.UserID.Hex(),
			Email:     userAction.Email,
			Action:    userAction.Action,
			Detail:    userAction.Detail,
			Category:  userAction.Category,
			CreatedAt: userAction.CreatedAt,
		}
		_, err = es.Client().Index().
			Index("user_actions").
			Id(userAction.UserID.Hex()).
			BodyJson(record).
			Do(context.Background())
		if err != nil {
			l.Logger.Fatal("[ERROR] restoring user_actions failed:", zap.Error(err))
			return
		}
		counter++
	}
	if err := cur.Err(); err != nil {
		l.Logger.Fatal("[ERROR] restoring user_actions failed:", zap.Error(err))
		return
	}
	cur.Close(context.TODO())

	l.Logger.Info(fmt.Sprintf("count %v", counter))
	l.Logger.Info(fmt.Sprintf("took  %v\n\n", time.Now().Sub(startTime)))
}
