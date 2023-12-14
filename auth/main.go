package main

import (
	"github.com/visonlv/iot-engine/auth/app"
	"github.com/visonlv/iot-engine/auth/model"
	"github.com/visonlv/iot-engine/auth/server"
)

func main() {
	app.Init("./config.toml")
	model.InitTable()
	server.Start()
}
