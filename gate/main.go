package main

import (
	"github.com/visonlv/iot-engine/gate/app"
	"github.com/visonlv/iot-engine/gate/server"
)

func main() {
	app.Init("")

	server.Start()
}
