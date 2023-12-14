package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/iot-engine/common/client/group"
	"github.com/visonlv/iot-engine/common/define"
	"github.com/visonlv/iot-engine/common/utils"
	grouppb "github.com/visonlv/iot-engine/group/proto"
	"github.com/visonlv/iot-engine/route/app"
	"github.com/visonlv/iot-engine/route/handler/client"
	"github.com/visonlv/iot-engine/route/handler/forwarding"
	"github.com/visonlv/iot-engine/route/server"
)

func handlerEvent(param *group.Param, data *grouppb.CategoryHeartBeatResp) error {
	return forwarding.ReloadClient(data)
}

func startConfigClient() {
	ip, err := utils.GetLocalIp()
	if err != nil {
		panic(err)
	}

	arr := strings.Split(app.Cfg.InnerClient.Group, ":")
	param := &group.Param{
		GroupIp:       arr[0],
		GroupPort:     arr[1],
		ServerIp:      ip,
		ServerPort:    fmt.Sprintf("%d", app.Cfg.Server.GrpcPort),
		RegisterCode:  define.CategoryRoute,
		SubscribeCode: define.CategoryShadow,
	}

	param.CallBack = func(data *grouppb.CategoryHeartBeatResp) {
		bb, _ := json.Marshal(data)
		logger.Infof("route callback msg:%s", string(bb))
		handlerEvent(param, data)
	}
	_, data, err := group.Start(param)
	if err != nil {
		panic(err)
	}
	handlerEvent(param, data)
}

func main() {
	// 1. 初始化配置
	app.Init("./config.toml")

	forwarding.Start()

	startConfigClient()

	go func() {
		time.Sleep(time.Second * 2)
		client.Start()
	}()

	server.Start()

}
