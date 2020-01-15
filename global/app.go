package global

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func init() {
	App.RootDir = "."
	if !viper.InConfig("http.port") {
		App.RootDir = inferRootDir()
	}
}

// App stores the information of the app.
var App = &app{}

type app struct {
	RootDir string
}

func exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

func inferRootDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	var infer func(d string) string
	infer = func(d string) string {
		if exist(d + "/config") {
			return d
		}
		return infer(filepath.Dir(d))
	}

	return infer(cwd)
}
