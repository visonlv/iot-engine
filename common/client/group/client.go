package group

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/visonlv/go-vkit/errorsx"
	"github.com/visonlv/go-vkit/grpcclient"
	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/go-vkit/utilsx"
	pb "github.com/visonlv/iot-engine/group/proto"
)

type Param struct {
	GroupIp       string
	GroupPort     string
	ServerIp      string
	ServerPort    string
	RegisterCode  string
	SubscribeCode string
	CallBack      func(*pb.CategoryHeartBeatResp)
}

type Client struct {
	LastVersion int64
	Data        *pb.CategoryHeartBeatResp
	Param       *Param
	GroupClient *pb.CategoryServiceClient

	Lock        *sync.Mutex
	Stream      *pb.CategoryService_HeartBeatClient
	StreamId    string
	CallBackMsg chan *pb.CategoryHeartBeatResp
}

func newClient(param *Param) *Client {
	addr := fmt.Sprintf("%s:%s", param.GroupIp, param.GroupPort)
	logger.Infof("[groupclient] group addr:%s", addr)
	cc := grpcclient.GetConnClient(
		addr,
		grpcclient.RequestTimeout(time.Second*10),
		grpcclient.DialTimeout(time.Second*10),
	)
	groupClient := pb.NewCategoryServiceClient("group", cc)

	client := &Client{
		LastVersion: -1,
		Data:        nil,
		Param:       param,
		GroupClient: groupClient,
		Lock:        new(sync.Mutex),
		CallBackMsg: make(chan *pb.CategoryHeartBeatResp, 128),
	}
	return client
}

func (c *Client) get() (*pb.CategoryHeartBeatResp, error) {
	resp, err := c.GroupClient.NodeList(context.Background(), &pb.CategoryNodeListReq{Code: c.Param.SubscribeCode})
	if err != nil {
		logger.Errorf("[groupclient] get fail:%s", err.Error())
		return nil, err
	}

	if resp.Code != 0 {
		logger.Errorf("[groupclient] get fail code:%d msg:%s", resp.Code, resp.Msg)
		return nil, fmt.Errorf("code:%d msg:%s", resp.Code, resp.Msg)
	}

	data := &pb.CategoryHeartBeatResp{
		Code:        resp.Code,
		Msg:         resp.Msg,
		LastVersion: resp.LastVersion,
		Items:       resp.Items,
	}

	if resp.Code != 0 {
		data.LastVersion = -1
		data.Items = make([]*pb.CategoryNodeItem, 0)
	}
	c.LastVersion = data.LastVersion
	c.Data = data
	return data, err
}

func (c *Client) mainloop() {
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	go c.run()

	// 五秒检查
	for {
		select {
		case <-t.C:
			c.Lock.Lock()
			if c.StreamId == "" {
				go c.run()
			}
			c.Lock.Unlock()
		case msg := <-c.CallBackMsg:
			c.Param.CallBack(msg)
		}
	}
}

func (c *Client) run() {
	streamId := utilsx.GenUuid()
	c.StreamId = streamId
	defer func() {
		logger.Infof("[groupclient] run defer StreamId:%v", c.StreamId)
		if streamId == c.StreamId {
			c.StreamId = ""
			c.Stream = nil
			// close
			data := &pb.CategoryHeartBeatResp{
				Code:        errorsx.FAIL.Code,
				Msg:         "关闭链接",
				LastVersion: -1,
				Items:       make([]*pb.CategoryNodeItem, 0),
			}
			c.LastVersion = data.LastVersion
			c.Data = data
			c.CallBackMsg <- data
		}
	}()

	stream, err := c.GroupClient.HeartBeat(context.Background())
	if err != nil {
		logger.Infof("[groupclient] create stream fail")
		return
	}

	c.Stream = stream

	stopCtx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(2)
	go c.readLoop(stopCtx, cancel, &wg)
	go c.writeLoop(stopCtx, cancel, &wg)
	wg.Wait()
}

func (c *Client) readLoop(stopCtx context.Context, cancelFunc context.CancelFunc, wg *sync.WaitGroup) {
	defer func() {
		cancelFunc()
		wg.Done()
	}()

	for {
		select {
		case <-stopCtx.Done():
			return
		default:
		}

		resp, err := c.Stream.Recv()
		if err == io.EOF || err != nil {
			logger.Errorf("[groupclient] readLoop StreamId:%s fail err:%v", c.StreamId, err)
			return
		}

		if resp.Code != 0 {
			resp.LastVersion = -1
			resp.Items = make([]*pb.CategoryNodeItem, 0)
		}
		c.LastVersion = resp.LastVersion
		c.Data = resp

		c.CallBackMsg <- resp
		if resp.Code != 0 {
			logger.Infof("注册节点失败,断开连接等待重连 code:%d msg:%s", resp.Code, resp.Msg)
			return
		}
	}
}

func (c *Client) writeLoop(stopCtx context.Context, cancelFunc context.CancelFunc, wg *sync.WaitGroup) {
	defer func() {
		cancelFunc()
		wg.Done()
	}()
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()

	err := c.heartBeat()
	if err != nil {
		return
	}

	for {
		select {
		case <-stopCtx.Done():
			return
		case <-t.C:
			err := c.heartBeat()
			if err != nil {
				return
			}

		}
	}
}

func (c *Client) heartBeat() error {
	err := c.Stream.Send(&pb.CategoryHeartBeatReq{
		RegisterCode:  c.Param.RegisterCode,
		SubscribeCode: c.Param.SubscribeCode,
		LastVersion:   c.LastVersion,
		Ip:            c.Param.ServerIp,
		Port:          c.Param.ServerPort,
	})
	if err != nil {
		logger.Infof("[groupclient] stream send fail %s", err.Error())
		return err
	}
	logger.Infof("[groupclient] stream send heartBeat success")
	return nil
}
