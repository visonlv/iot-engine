package stream

import (
	"fmt"
	"time"

	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/go-vkit/utilsx"
	"github.com/visonlv/iot-engine/group/model"
	pb "github.com/visonlv/iot-engine/group/proto"
)

var _p *Stream = nil

func Start() {
	_p = newStream()
	err := _p.start()
	if err != nil {
		panic(err)
	}
}

func CreateSession(s *pb.CategoryService_HeartBeatServer, clientId string, clientInfo *ClientInfo) *StreamServiceClient {
	c := &StreamServiceClient{
		S:           s,
		Stop:        make(chan struct{}),
		Msgs:        make(chan *pb.CategoryHeartBeatResp, QUEUE_SIZE),
		ClientId:    clientId,
		SessionId:   utilsx.GenUuid(),
		ClientInfo:  clientInfo,
		LastRevTime: time.Now().Unix(),
	}
	return c
}

func RunClient(c *StreamServiceClient, msg *pb.CategoryHeartBeatReq) {
	logger.Infof("[stream] RunClient clientId:%s sessionId:%s 上线", c.ClientId, c.SessionId)
	_p.h.kickoutAndAddClient(c)
	_p.run(c, msg)
	logger.Infof("[stream] RunClient clientId:%s sessionId:%s 下线 ", c.ClientId, c.SessionId)
}

func NodeListByCode(code string) (list []*pb.CategoryNodeItem, lastVersion int64, err error) {
	subscribeInfo, subscribeOk := _p.h.getCategoryByCode(code)
	if !subscribeOk {
		return make([]*pb.CategoryNodeItem, 0), 0, nil
	} else {
		return subscribeInfo.List, subscribeInfo.LastVersion, nil
	}
}

func GetClientId(ip, port string) string {
	return fmt.Sprintf("%s:%s", ip, port)
}

func ReloadCategory(categoryModel *model.CategoryModel) {
	_p.h.reloadCategory(categoryModel)
}
