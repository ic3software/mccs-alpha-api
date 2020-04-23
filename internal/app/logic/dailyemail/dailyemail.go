package dailyemail

import (
	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/email"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/helper"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/l"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Run performs the daily email notification.
func Run() {
	users, err := logic.User.FindByDailyNotification()
	if err != nil {
		l.Logger.Error("dailyemail failed", zap.Error(err))
		return
	}

	viper.SetDefault("concurrency_num", 1)
	pool := NewPool(viper.GetInt("concurrency_num"))

	for _, user := range users {
		worker := createEmailWorker(user)
		pool.Run(worker)
	}

	pool.Shutdown()
}

func createEmailWorker(u *types.User) func() {
	return func() {
		matchedTags, err := getMatchTags(u)
		if err != nil {
			l.Logger.Error("dailyemail failed", zap.Error(err))
			return
		}
		if len(matchedTags.MatchedOffers) == 0 && len(matchedTags.MatchedWants) == 0 {
			return
		}
		err = email.SendDailyEmailList(u, matchedTags)
		if err != nil {
			l.Logger.Error("dailyemail failed", zap.Error(err))
		}
		err = logic.User.UpdateLastNotificationSentDate(u.ID)
		if err != nil {
			l.Logger.Error("dailyemail failed", zap.Error(err))
		}
	}
}

func getMatchTags(user *types.User) (*types.MatchedTags, error) {
	entity, err := logic.Entity.FindByID(user.Entities[0])
	if err != nil {
		return nil, err
	}

	matchedOffers, err := logic.Tag.MatchOffers(helper.GetTagNames(entity.Offers), user.LastNotificationSentDate)
	if err != nil {
		return nil, err
	}
	matchedWants, err := logic.Tag.MatchWants(helper.GetTagNames(entity.Wants), user.LastNotificationSentDate)
	if err != nil {
		return nil, err
	}

	return &types.MatchedTags{
		MatchedOffers: matchedOffers,
		MatchedWants:  matchedWants,
	}, nil
}
