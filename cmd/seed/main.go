package main

import (
	"github.com/ic3network/mccs-alpha-api/global"
	"github.com/ic3network/mccs-alpha-api/internal/seed"
)

func main() {
	global.Init()
	seed.LoadData()
	seed.Run()
}
