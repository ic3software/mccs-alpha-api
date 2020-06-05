package dailyemail

import (
	"time"

	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/email"
	"github.com/ic3network/mccs-alpha-api/util"
	"github.com/ic3network/mccs-alpha-api/util/l"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// Run performs the daily email notification.
func Run() {
	entities, err := logic.Entity.FindByDailyNotification()
	if err != nil {
		l.Logger.Error("dailyemail failed", zap.Error(err))
		return
	}

	viper.SetDefault("concurrency_num", 1)
	pool := NewPool(viper.GetInt("concurrency_num"))

	for _, entity := range entities {
		worker := createEmailWorker(entity)
		pool.Run(worker)
	}

	pool.Shutdown()
}

func createEmailWorker(entity *types.Entity) func() {
	return func() {
		if !util.IsAcceptedStatus(entity.Status) {
			return
		}

		matchedTags, err := getMatchTags(entity, entity.LastNotificationSentDate)
		if err != nil {
			l.Logger.Error("dailyemail failed", zap.Error(err))
			return
		}

		if len(matchedTags.MatchedOffers) == 0 && len(matchedTags.MatchedWants) == 0 {
			return
		}

		err = email.SendDailyEmailList(entity, matchedTags, entity.LastNotificationSentDate)
		if err != nil {
			l.Logger.Error("dailyemail failed", zap.Error(err))
		}

		err = logic.Entity.UpdateLastNotificationSentDate(entity.ID)
		if err != nil {
			l.Logger.Error("dailyemail failed", zap.Error(err))
		}
	}
}

func getEntity(entityID primitive.ObjectID) (*types.Entity, error) {
	entity, err := logic.Entity.FindByID(entityID)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func getMatchTags(entity *types.Entity, lastNotificationSentDate time.Time) (*types.MatchedTags, error) {
	matchedOffers, err := logic.Tag.MatchOffers(types.TagFieldToNames(entity.Offers), lastNotificationSentDate)
	if err != nil {
		return nil, err
	}
	matchedWants, err := logic.Tag.MatchWants(types.TagFieldToNames(entity.Wants), lastNotificationSentDate)
	if err != nil {
		return nil, err
	}
	return &types.MatchedTags{
		MatchedOffers: matchedOffers,
		MatchedWants:  matchedWants,
	}, nil
}
