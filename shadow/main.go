package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/iot-engine/common/client/group"
	"github.com/visonlv/iot-engine/common/define"
	"github.com/visonlv/iot-engine/common/utils"
	grouppb "github.com/visonlv/iot-engine/group/proto"
	"github.com/visonlv/iot-engine/shadow/app"
	"github.com/visonlv/iot-engine/shadow/handler/client"
	"github.com/visonlv/iot-engine/shadow/handler/forwarding"
	"github.com/visonlv/iot-engine/shadow/model"
	"github.com/visonlv/iot-engine/shadow/proxy"
	"github.com/visonlv/iot-engine/shadow/server"
)

var nats2LocalProxy *proxy.Nats2Local

func handlerEvent(param *group.Param, data *grouppb.CategoryHeartBeatResp) bool {
	//取消所有订阅
	if data.Code != 0 {
		logger.Infof("handlerEvent client stream error code:%d, msg:%s", data.Code, data.Msg)
		nats2LocalProxy.ReloadGroup(-1, -1)
		app.Cfg.Group.Start = -1
		app.Cfg.Group.End = -1
		return false
	}
	for _, v := range data.Items {
		if v.Port == param.ServerPort && v.Ip == param.ServerIp {
			if v.Status == 1 {
				nats2LocalProxy.ReloadGroup(v.Start, v.End)
				app.Cfg.Group.Start = v.Start
				app.Cfg.Group.End = v.End
				return true
			}
		}
	}
	return false
}

func startProxy() {
	nats2LocalProxy = proxy.NewNats2Local("Nats2Local", "iot.up", "local", "")
	err := nats2LocalProxy.Start(&forwarding.LocalPub{})
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
		RegisterCode:  define.CategoryShadow,
		SubscribeCode: define.CategoryShadow,
	}

	hitGroup := make(chan struct{})
	var isClosed int32 = 0
	param.CallBack = func(data *grouppb.CategoryHeartBeatResp) {
		bb, _ := json.Marshal(data)
		logger.Infof("shadow callback msg:%s", string(bb))
		hit := handlerEvent(param, data)
		if hit && atomic.LoadInt32(&isClosed) == 0 {
			atomic.StoreInt32(&isClosed, 1)
			close(hitGroup)
		}
	}
	group.Start(param)
	go func() {
		time.Sleep(time.Second * 20)
		if atomic.LoadInt32(&isClosed) == 0 {
			panic("register node time out")
		}
	}()
	<-hitGroup
}

func initTDengineTable() {
	model.CreateDeviceMsgLogStable(context.Background())
}

func main() {
	// 1. 初始化配置
	app.Init("./config.toml")

	model.InitTable()

	initTDengineTable()

	startProxy()

	forwarding.Start()

	startConfigClient()

	go func() {
		time.Sleep(time.Second * 2)
		client.Start()

	}()

	server.Start()

}
