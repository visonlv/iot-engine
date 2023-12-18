package main

import (
	"github.com/visonlv/iot-engine/notify/app"
	"github.com/visonlv/iot-engine/notify/model"
	"github.com/visonlv/iot-engine/notify/server"
)

func main() {
	// 1. 初始化配置
	app.Init("./config.toml")

	model.InitTable()

	server.Start()

}
