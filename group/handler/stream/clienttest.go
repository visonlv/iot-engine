package stream

import (
	"encoding/json"

	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/iot-engine/common/client/group"
	"github.com/visonlv/iot-engine/common/define"
	pb "github.com/visonlv/iot-engine/group/proto"
)

func route1() {
	param := &group.Param{
		GroupIp:       "127.0.0.1",
		GroupPort:     "10003",
		ServerIp:      "127.0.0.1",
		ServerPort:    "7001",
		RegisterCode:  define.CategoryRoute,
		SubscribeCode: define.CategoryShadow,
		CallBack: func(chbr *pb.CategoryHeartBeatResp) {
			bb, _ := json.Marshal(chbr)
			logger.Infof("route1 7001 callback msg:%s", string(bb))
		},
	}
	group.Start(param)
}

func route2() {
	param := &group.Param{
		GroupIp:       "127.0.0.1",
		GroupPort:     "10003",
		ServerIp:      "127.0.0.1",
		ServerPort:    "7002",
		RegisterCode:  define.CategoryRoute,
		SubscribeCode: define.CategoryShadow,
		CallBack: func(chbr *pb.CategoryHeartBeatResp) {
			bb, _ := json.Marshal(chbr)
			logger.Infof("route2 7002 callback msg:%s", string(bb))
		},
	}
	group.Start(param)
}

func shadow1() {
	param := &group.Param{
		GroupIp:       "127.0.0.1",
		GroupPort:     "10003",
		ServerIp:      "127.0.0.1",
		ServerPort:    "8001",
		RegisterCode:  define.CategoryShadow,
		SubscribeCode: define.CategoryShadow,
		CallBack: func(chbr *pb.CategoryHeartBeatResp) {
			bb, _ := json.Marshal(chbr)
			logger.Infof("shadow1 80001 callback msg:%s", string(bb))
		},
	}
	group.Start(param)
}

func shadow2() {
	param := &group.Param{
		GroupIp:       "127.0.0.1",
		GroupPort:     "10003",
		ServerIp:      "127.0.0.1",
		ServerPort:    "8002",
		RegisterCode:  define.CategoryShadow,
		SubscribeCode: define.CategoryShadow,
		CallBack: func(chbr *pb.CategoryHeartBeatResp) {
			bb, _ := json.Marshal(chbr)
			logger.Infof("shadow2 8002 callback msg:%s", string(bb))
		},
	}
	group.Start(param)
}

func proxy1() {
	param := &group.Param{
		GroupIp:       "127.0.0.1",
		GroupPort:     "10003",
		ServerIp:      "127.0.0.1",
		ServerPort:    "9001",
		RegisterCode:  define.CategoryProxy,
		SubscribeCode: define.CategoryProxy,
		CallBack: func(chbr *pb.CategoryHeartBeatResp) {
			bb, _ := json.Marshal(chbr)
			logger.Infof("proxy1 9001 callback msg:%s", string(bb))
		},
	}
	group.Start(param)
}

func proxy2() {
	param := &group.Param{
		GroupIp:       "127.0.0.1",
		GroupPort:     "10003",
		ServerIp:      "127.0.0.1",
		ServerPort:    "9002",
		RegisterCode:  define.CategoryProxy,
		SubscribeCode: define.CategoryProxy,
		CallBack: func(chbr *pb.CategoryHeartBeatResp) {
			bb, _ := json.Marshal(chbr)
			logger.Infof("proxy2 9002 callback msg:%s", string(bb))
		},
	}
	group.Start(param)
}

func StartClient() {
	shadow1()
	shadow2()

	// proxy1()
	// proxy2()

	// route1()
	// route2()
}
