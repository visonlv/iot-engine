package main

import (
	"github.com/visonlv/iot-engine/rule/app"
	"github.com/visonlv/iot-engine/rule/model"
	"github.com/visonlv/iot-engine/rule/server"
)

func main() {
	// 1. 初始化配置
	app.Init("./config.toml")

	model.InitTable()

	server.Start()

}
