package stream

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/iot-engine/common/define"
	pb "github.com/visonlv/iot-engine/group/proto"
)

type ClientInfo struct {
	ClientId      string
	RegisterCode  string
	SubscribeCode string
}

type StreamServiceClient struct {
	S           *pb.CategoryService_HeartBeatServer
	Msgs        chan *pb.CategoryHeartBeatResp
	Stop        chan struct{}
	ClientId    string
	SessionId   string
	ClientInfo  *ClientInfo
	LastRevTime int64
}

type CategoryInfo struct {
	Code        string
	LastVersion int64
	Lock        *sync.RWMutex
	List        []*pb.CategoryNodeItem
	Map         map[string]*pb.CategoryNodeItem
}

func (c *CategoryInfo) packMsg() *pb.CategoryHeartBeatResp {
	dst := make([]*pb.CategoryNodeItem, len(c.List))
	copy(dst, c.List)
	return &pb.CategoryHeartBeatResp{
		Code:        0,
		Msg:         "",
		LastVersion: c.LastVersion,
		Items:       dst,
	}
}

func (c *CategoryInfo) packErrMsg(code int32, msg string) *pb.CategoryHeartBeatResp {
	return &pb.CategoryHeartBeatResp{
		Code: code,
		Msg:  msg,
	}
}

func (c *CategoryInfo) getByClientId(clientId string) (*pb.CategoryNodeItem, bool) {
	v, ok := c.Map[clientId]
	return v, ok
}

func (c *CategoryInfo) addNode(node *pb.CategoryNodeItem) *pb.CategoryNodeItem {
	clientId := GetClientId(node.Ip, node.Port)
	_, ok := c.Map[clientId]
	if ok {
		panic(fmt.Sprintf("[stream] addNode fail clientId:%s exist", clientId))

	}
	c.Map[clientId] = node
	c.List = append(c.List, node)
	c.LastVersion++

	bb, _ := json.Marshal(node)
	logger.Infof("[stream] addNode addNode success index:%s clientId:%s data:%s", node.Index, clientId, string(bb))
	return node
}

func (c *CategoryInfo) removeNode(clientId string) bool {
	v, ok := c.Map[clientId]
	if ok {
		delete(c.Map, clientId)
		for index, v2 := range c.List {
			if v2.Ip == v.Ip && v2.Port == v.Port {
				c.List = append(c.List[:index], c.List[index+1:]...)
			}
		}
		c.LastVersion++

		bb, _ := json.Marshal(v)
		logger.Infof("[stream] removeNode success index:%s clientId:%s data:%s", v.Index, clientId, string(bb))
		return true
	}
	return false
}

func (c *CategoryInfo) bingNode(ip, port string) (*pb.CategoryNodeItem, error) {
	clientId := GetClientId(ip, port)
	_, ok := c.Map[clientId]
	if ok {
		panic(fmt.Sprintf("[stream] bingNode fail clientId:%s exist", clientId))
	}
	var hitNode *pb.CategoryNodeItem
	for _, v := range c.List {
		if v.Status == define.NodeUnBind {
			hitNode = v
			break
		}
	}
	if hitNode == nil {
		return nil, fmt.Errorf("预设分组已经都被占用")
	}

	// 占用一个
	hitNode.Ip = ip
	hitNode.Port = port
	hitNode.Status = define.NodeBind
	c.Map[clientId] = hitNode
	c.LastVersion++

	bb, _ := json.Marshal(hitNode)
	logger.Infof("[stream] bingNode success index:%s clientId:%s data:%s", hitNode.Index, clientId, string(bb))

	return hitNode, nil
}

func (c *CategoryInfo) unbingNode(clientId string) bool {
	v, ok := c.Map[clientId]
	if ok {
		delete(c.Map, clientId)
		v.Ip = ""
		v.Port = ""
		v.Status = define.NodeUnBind
		c.LastVersion++

		bb, _ := json.Marshal(v)
		logger.Infof("[stream] unbingNode success index:%s clientId:%s data:%s", v.Index, clientId, string(bb))
		return true
	}
	return false
}
