package global

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	once   = new(sync.Once)
	config string
)

// Init initiates the app configurations.
func Init() {
	once.Do(func() {
		config = os.Getenv("CONFIG")
		if config == "" {
			config = "development"
		}
		if err := initConfig(); err != nil {
			panic(fmt.Errorf("init config failed: %s", err))
		}
		watchConfig()
	})
}

func initConfig() error {
	viper.SetConfigName(config)
	viper.AddConfigPath("config")
	viper.AddConfigPath(App.RootDir + "/config")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	return nil
}

func watchConfig() {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Printf("Config file changed: %s", e.Name)
	})
}
