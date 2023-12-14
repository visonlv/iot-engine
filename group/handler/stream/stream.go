package stream

import (
	"context"
	"fmt"
	"io"
	"runtime/debug"
	"sync"
	"time"

	"github.com/visonlv/go-vkit/errorsx"
	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/iot-engine/common/define"
	pb "github.com/visonlv/iot-engine/group/proto"
)

const (
	QUEUE_SIZE = 128
)

type Stream struct {
	h *Hub
}

func newStream() *Stream {
	return &Stream{
		h: newHub(),
	}
}

func (s *Stream) start() error {
	return s.h.loadFromDb()
}

func (s *Stream) run(c *StreamServiceClient, msg *pb.CategoryHeartBeatReq) {
	defer func() {
		logger.Infof("[stream] Run defer clientId:%s sessionId:%s", c.ClientId, c.SessionId)
	}()
	stopCtx, cancel := context.WithCancel(context.Background())
	go s.readLoop(stopCtx, cancel, c)
	go s.writeLoop(stopCtx, cancel, c)
	// 处理第一个消息
	err := s.handleFirstMsg(c, msg)
	if err != nil {
		logger.Errorf("[stream] handleMsg clientId:%s sessionId:%s fail err:%v", c.ClientId, c.SessionId, err)
	}

	<-c.Stop
	cancel()
}

func (s *Stream) readLoop(stopCtx context.Context, cancelFunc context.CancelFunc, c *StreamServiceClient) {
	defer func() {
		if r := recover(); r != nil {
			errorStr := fmt.Sprintf("[stream] readLoop clientId:%s sessionId:%s panic recovered:%v ", c.ClientId, c.SessionId, r)
			logger.Errorf(errorStr)
			logger.Error(string(debug.Stack()))
		}
	}()

	defer func() {
		cancelFunc()
		s.h.removeByClient(c)
		logger.Infof("[stream] readLoop defer clientId:%s sessionId:%s", c.ClientId, c.SessionId)
	}()

	for {
		select {
		case <-stopCtx.Done():
			return
		default:
		}

		msg, err := c.S.Recv()
		if err == io.EOF || err != nil {
			logger.Errorf("[stream] readLoop clientId:%s sessionId:%s fail err:%v", c.ClientId, c.ClientId, err)
			return
		}
		c.LastRevTime = time.Now().Unix()
		logger.Infof("[stream] readLoop clientId:%s sessionId:%s version:%d", c.ClientId, c.ClientId, msg.LastVersion)

		err = s.handleMsg(c, msg)
		if err != nil {
			logger.Errorf("[stream] handleMsg clientId:%s sessionId:%s fail err:%v", c.ClientId, c.ClientId, err)
		}
	}
}

func (s *Stream) writeLoop(stopCtx context.Context, cancelFunc context.CancelFunc, c *StreamServiceClient) {
	defer func() {
		cancelFunc()
		s.h.removeByClient(c)
		logger.Infof("[stream] writeLoop defer clientId:%s sessionId:%s", c.ClientId, c.SessionId)
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stopCtx.Done():
			return
		case <-ticker.C:
			if time.Now().Unix()-c.LastRevTime >= 10 {
				logger.Infof("[stream] writeLoop LastRevTime timeout clientId:%s sessionId:%s", c.ClientId, c.SessionId)
				return
			}
		case msg := <-c.Msgs:
			err := c.S.Send(msg)
			if err != nil {
				logger.Errorf("[stream] writeLoop Send clientId:%s sessionId:%s err:%v", c.ClientId, c.SessionId, err)
				return
			}
		}
	}
}

func (s *Stream) handleFirstMsg(c *StreamServiceClient, msg *pb.CategoryHeartBeatReq) error {
	var err error
	if err = msg.Validate(); err != nil {
		logger.Errorf("[stream] handleFirstMsg clientId:%s sessionId:%s Validate err:%v", c.ClientId, c.SessionId, err)
		return err
	}
	clientId := GetClientId(msg.Ip, msg.Port)

	// 注册分类
	s.h.CategoryLock.Lock()
	registerCode := msg.RegisterCode
	registerInfo, registerOk := s.h.getCategoryByCode(registerCode)

	//注册逻辑
	//1、路由节点可以新建分类
	if !registerOk && registerCode == define.CategoryRoute {
		registerInfo = &CategoryInfo{
			List:        make([]*pb.CategoryNodeItem, 0),
			Map:         make(map[string]*pb.CategoryNodeItem),
			LastVersion: 0,
			Lock:        new(sync.RWMutex),
			Code:        registerCode,
		}
		s.h.addCategory(registerInfo)
	}
	s.h.CategoryLock.Unlock()

	if registerInfo == nil {
		logger.Errorf("[stream] handleFirstMsg clientId:%s sessionId:%s registerInfo registerCode:%v", c.ClientId, c.SessionId, registerCode)
		return fmt.Errorf("not support registerCode:%s", registerCode)
	}

	registerInfo.Lock.Lock()

	oldVersion := registerInfo.LastVersion
	nodeInfo, ok := registerInfo.getByClientId(clientId)
	//2、节点不存在 开始注册节点
	if !ok {
		//注册新节点
		if registerCode == define.CategoryRoute {
			nodeInfo = &pb.CategoryNodeItem{
				Index:  GetClientId(msg.Ip, msg.Port),
				Ip:     msg.Ip,
				Port:   msg.Port,
				Status: define.NodeBind,
			}
			nodeInfo = registerInfo.addNode(nodeInfo)
		} else {
			//绑定节点
			nodeInfo, err = registerInfo.bingNode(msg.Ip, msg.Port)
			if err != nil {
				registerInfo.Lock.Unlock()
				s.h.sendByClientId(c.ClientId, registerInfo.packErrMsg(errorsx.FAIL.Code, err.Error()))
				return err
			}
		}
	}
	registerInfo.Lock.Unlock()

	if oldVersion != registerInfo.LastVersion {
		s.h.broadcast(registerInfo.Code)
	}
	return nil
}

func (s *Stream) handleMsg(c *StreamServiceClient, msg *pb.CategoryHeartBeatReq) error {
	// 订阅分类
	subscribeCode := msg.SubscribeCode
	s.h.CategoryLock.RLock()
	subscribeInfo, subscribeOk := s.h.getCategoryByCode(subscribeCode)
	s.h.CategoryLock.RUnlock()
	sendMsg := &pb.CategoryHeartBeatResp{}
	if !subscribeOk {
		sendMsg = &pb.CategoryHeartBeatResp{
			Code:        0,
			Msg:         fmt.Sprintf("not support subscribeCode:%s", subscribeCode),
			LastVersion: 0,
			Items:       make([]*pb.CategoryNodeItem, 0),
		}
	} else {
		subscribeInfo.Lock.RLock()
		sendMsg = subscribeInfo.packMsg()
		subscribeInfo.Lock.RUnlock()
	}

	if msg.LastVersion != sendMsg.LastVersion {
		return s.h.sendByClientId(c.ClientId, sendMsg)
	}
	return nil
}
