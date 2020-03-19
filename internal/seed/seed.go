package seed

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"time"

	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/bcrypt"
)

var (
	entityData       []types.Entity
	balanceLimitData []types.BalanceLimit
	userData         []types.User
	adminUserData    []types.AdminUser
	tagData          []types.Tag
	categoriesData   []types.Category
)

func LoadData() {
	// Load entity data.
	data, err := ioutil.ReadFile("internal/seed/data/entity.json")
	if err != nil {
		log.Fatal(err)
	}
	entities := make([]types.Entity, 0)
	json.Unmarshal(data, &entities)
	entityData = entities

	// Load balance limit data.
	data, err = ioutil.ReadFile("internal/seed/data/balance_limits.json")
	if err != nil {
		log.Fatal(err)
	}
	balanceLimits := make([]types.BalanceLimit, 0)
	json.Unmarshal(data, &balanceLimits)
	balanceLimitData = balanceLimits

	// Load user data.
	data, err = ioutil.ReadFile("internal/seed/data/user.json")
	if err != nil {
		log.Fatal(err)
	}
	users := make([]types.User, 0)
	json.Unmarshal(data, &users)
	userData = users

	// Load admin user data.
	data, err = ioutil.ReadFile("internal/seed/data/admin_user.json")
	if err != nil {
		log.Fatal(err)
	}
	adminUsers := make([]types.AdminUser, 0)
	json.Unmarshal(data, &adminUsers)
	adminUserData = adminUsers

	// Load user tag data.
	data, err = ioutil.ReadFile("internal/seed/data/tag.json")
	if err != nil {
		log.Fatal(err)
	}
	tags := make([]types.Tag, 0)
	json.Unmarshal(data, &tags)
	tagData = tags

	// Load admin tag data.
	data, err = ioutil.ReadFile("internal/seed/data/category.json")
	if err != nil {
		log.Fatal(err)
	}
	categories := make([]types.Category, 0)
	json.Unmarshal(data, &categories)
	categoriesData = categories
}

func Run() {
	log.Println("start seeding")
	startTime := time.Now()

	// create users and entities.
	for i, b := range entityData {
		accountNumber, err := PostgresSQL.CreateAccount()
		if err != nil {
			log.Fatal(err)
		}
		b.AccountNumber = accountNumber

		entityID, err := MongoDB.CreateEntity(b)
		if err != nil {
			log.Fatal(err)
		}
		b.ID = entityID

		err = ElasticSearch.CreateEntity(&b)
		if err != nil {
			log.Fatal(err)
		}

		u := userData[i]
		u.Entities = append(u.Entities, b.ID)
		hashedPassword, _ := bcrypt.Hash(u.Password)
		u.Password = hashedPassword

		userID, err := MongoDB.CreateUser(u)
		if err != nil {
			log.Fatal(err)
		}
		u.ID = userID

		err = MongoDB.AssociateUserWithEntity(u.ID, b.ID)
		if err != nil {
			log.Fatal(err)
		}

		err = ElasticSearch.CreateUser(&u)
		if err != nil {
			log.Fatal(err)
		}
	}

	err := PostgresSQL.UpdateBalanceLimits(balanceLimitData)
	if err != nil {
		log.Fatal(err)
	}

	err = MongoDB.CreateAdminUsers(adminUserData)
	if err != nil {
		log.Fatal(err)
	}

	err = MongoDB.CreateTags(tagData)
	if err != nil {
		log.Fatal(err)
	}

	err = MongoDB.CreateCategories(categoriesData)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("took  %v\n", time.Now().Sub(startTime))
}
