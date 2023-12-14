package main

import (
	"github.com/visonlv/iot-engine/thing/app"
	"github.com/visonlv/iot-engine/thing/handler/product"
	"github.com/visonlv/iot-engine/thing/model"
	"github.com/visonlv/iot-engine/thing/server"
)

func main() {
	// 1. 初始化配置
	app.Init("./config.toml")

	model.InitTable()

	product.InitAllProduct()

	server.Start()

}
