package main

import (
	"github.com/ic3network/mccs-alpha-api/global"
	"github.com/ic3network/mccs-alpha-api/internal/app/http"
	"github.com/ic3network/mccs-alpha-api/internal/app/logic/balancecheck"
	"github.com/ic3network/mccs-alpha-api/internal/app/logic/dailyemail"
	"github.com/ic3network/mccs-alpha-api/util/l"
	"github.com/robfig/cron"
	"github.com/spf13/viper"
)

func init() {
	global.Init()
}

func main() {
	// Flushes log buffer, if any.
	defer l.Logger.Sync()
	go ServeBackGround()
	go RunMigration()

	http.AppServer.Run(viper.GetString("port"))
}

// ServeBackGround performs the background activities.
func ServeBackGround() {
	c := cron.New()
	viper.SetDefault("daily_email_schedule", "0 0 7 * * *")
	c.AddFunc(viper.GetString("daily_email_schedule"), func() {
		l.Logger.Info("[ServeBackGround] Running daily email schedule. \n")
		dailyemail.Run()
	})
	viper.SetDefault("balance_check_schedule", "0 0 * * * *")
	c.AddFunc(viper.GetString("balance_check_schedule"), func() {
		l.Logger.Info("[ServeBackGround] Running balance check schedule. \n")
		balancecheck.Run()
	})
	c.Start()
}

func RunMigration() {
}
