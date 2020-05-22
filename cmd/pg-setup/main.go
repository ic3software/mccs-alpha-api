package main

import (
	"github.com/ic3network/mccs-alpha-api/global"

	"context"
	"log"
	"time"

	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/mongo"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	global.Init()
	// setUpAccount()
}

// setUpAccount reads the entities from MongoDB and build up the accounts in PostgreSQL.
func setUpAccount() {
	log.Println("start setting up accounts in PostgreSQL")
	startTime := time.Now()
	ctx := context.Background()

	filter := bson.M{
		"deletedAt": bson.M{"$exists": false},
	}
	cur, err := mongo.DB().Collection("entities").Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}

	counter := 0
	for cur.Next(ctx) {
		var entity types.Entity
		err := cur.Decode(&entity)
		if err != nil {
			log.Fatal(err)
		}
		// Create account from entity.
		account, err := logic.Account.Create()
		if err != nil {
			log.Fatal(err)
		}
		err = updateEntity(&entity, account)
		if err != nil {
			log.Fatal(err)
		}
		counter++
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	cur.Close(ctx)

	log.Printf("count %v\n", counter)
	log.Printf("took  %v\n\n", time.Now().Sub(startTime))
}

func updateEntity(entity *types.Entity, account *types.Account) error {
	filter := bson.M{"_id": entity.ID}
	update := bson.M{"$set": bson.M{
		"accountNumber": account.AccountNumber,
		"updatedAt":     time.Now(),
	}}

	_, err := mongo.DB().Collection("entities").UpdateOne(
		context.Background(),
		filter,
		update,
	)
	if err != nil {
		return err
	}

	return nil
}
