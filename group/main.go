package main

import (
	"time"

	"github.com/visonlv/iot-engine/group/app"
	"github.com/visonlv/iot-engine/group/handler/stream"
	"github.com/visonlv/iot-engine/group/model"
	"github.com/visonlv/iot-engine/group/server"
)

func main() {
	// 1. 初始化配置
	app.Init("./config.toml")

	model.InitTable()

	stream.Start()
	go func() {
		time.Sleep(time.Second * 3)
		// stream.StartClient()
	}()

	server.Start()

}
