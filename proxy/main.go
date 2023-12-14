package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/iot-engine/common/client/group"
	"github.com/visonlv/iot-engine/common/define"
	"github.com/visonlv/iot-engine/common/utils"
	grouppb "github.com/visonlv/iot-engine/group/proto"
	"github.com/visonlv/iot-engine/proxy/app"
	"github.com/visonlv/iot-engine/proxy/proxy"
)

var emqx2NatsProxy *proxy.Emqx2Nats
var nats2EmqxProxy *proxy.Nats2Emqx

func handlerEvent(param *group.Param, data *grouppb.CategoryHeartBeatResp) {
	//取消所有订阅
	if data.Code != 0 {
		logger.Infof("handlerEvent client stream error code:%d, msg:%s", data.Code, data.Msg)
		emqx2NatsProxy.ReloadGroup(-1, -1)
		nats2EmqxProxy.ReloadGroup(-1, -1)
		return
	}
	for _, v := range data.Items {
		if v.Port == param.ServerPort && v.Ip == param.ServerIp {
			if v.Status == 1 {
				emqx2NatsProxy.ReloadGroup(v.Start, v.End)
				nats2EmqxProxy.ReloadGroup(v.Start, v.End)
				return
			}
		}
	}
}

func startProxy() {
	emqx2NatsProxy = proxy.NewEmqx2Nats("Emqx2Nats", "iot/up", "nats", "")
	nats2EmqxProxy = proxy.NewNats2Emqx("Nats2Emqx", "iot.down", "emqx", "")
	err := emqx2NatsProxy.Start()
	if err != nil {
		panic(err)
	}
	err = nats2EmqxProxy.Start()
	if err != nil {
		panic(err)
	}
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
		RegisterCode:  define.CategoryProxy,
		SubscribeCode: define.CategoryProxy,
	}
	param.CallBack = func(data *grouppb.CategoryHeartBeatResp) {
		bb, _ := json.Marshal(data)
		logger.Infof("proxy1 9001 callback msg:%s", string(bb))
		handlerEvent(param, data)
	}
	_, data, err := group.Start(param)
	if err != nil {
		panic(err)
	}
	if data.Code != 0 || len(data.Items) == 0 {
		panic("not config")
	}
	handlerEvent(param, data)
}

func main() {
	// 1. 初始化配置
	app.Init("./config.toml")

	startProxy()

	startConfigClient()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
