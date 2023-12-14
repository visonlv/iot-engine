package stream

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/iot-engine/common/define"
	"github.com/visonlv/iot-engine/group/handler/category"
	"github.com/visonlv/iot-engine/group/model"
	pb "github.com/visonlv/iot-engine/group/proto"
)

type Hub struct {
	ClientLock     *sync.RWMutex
	CategoryLock   *sync.RWMutex
	CategoryMap    map[string]*CategoryInfo
	Cid2Client     map[string]*StreamServiceClient
	Session2Client map[string]*StreamServiceClient
}

func newHub() *Hub {
	return &Hub{
		ClientLock:     new(sync.RWMutex),
		CategoryLock:   new(sync.RWMutex),
		CategoryMap:    make(map[string]*CategoryInfo),
		Cid2Client:     make(map[string]*StreamServiceClient),
		Session2Client: make(map[string]*StreamServiceClient),
	}
}

func (h *Hub) kickoutAndAddClient(c *StreamServiceClient) {
	h.ClientLock.Lock()
	var delClient *StreamServiceClient
	if tempC, ok := h.Cid2Client[c.ClientId]; ok {
		delete(h.Cid2Client, tempC.ClientId)
		delete(h.Session2Client, tempC.SessionId)
		delClient = tempC
	}
	if delClient != nil {
		close(delClient.Stop)
		logger.Infof("[stream] kickoutAndAddClient clientId:%s sessionId:%s remove success", delClient.ClientId, delClient.SessionId)
	}

	h.Cid2Client[c.ClientId] = c
	h.Session2Client[c.SessionId] = c
	logger.Infof("[stream] kickoutAndAddClient clientId:%s sessionId:%s add success", c.ClientId, c.SessionId)
	h.ClientLock.Unlock()

	//移除缓存节点
	if delClient != nil {
		h.removeNodeAndBroadcast(delClient.ClientInfo)
	}
}

func (h *Hub) removeByClient(c *StreamServiceClient) {
	h.ClientLock.Lock()
	var delClient *StreamServiceClient
	if tempC, ok := h.Session2Client[c.SessionId]; ok {
		delete(h.Session2Client, tempC.SessionId)
		delete(h.Cid2Client, tempC.ClientId)
		delClient = tempC
	}
	h.ClientLock.Unlock()

	if delClient == nil {
		return
	}
	close(delClient.Stop)
	logger.Infof("[stream] removeByClient clientId:%s sessionId:%s", delClient.ClientId, delClient.SessionId)
	//移除缓存节点
	h.removeNodeAndBroadcast(delClient.ClientInfo)
}

func (h *Hub) _removeByClientId(clientId string) {
	h.ClientLock.Lock()
	defer h.ClientLock.Unlock()

	var delClient *StreamServiceClient
	if tempC, ok := h.Cid2Client[clientId]; ok {
		delete(h.Cid2Client, tempC.ClientId)
		delete(h.Session2Client, tempC.SessionId)
		delClient = tempC
	}

	if delClient == nil {
		return
	}

	close(delClient.Stop)
	logger.Infof("[stream] _removeByClientId ClientId:%v clientAddr:%p", delClient.ClientId, delClient)
}

func (h *Hub) sendByClientId(clientId string, resp *pb.CategoryHeartBeatResp) error {
	bb, _ := json.Marshal(resp)

	h.ClientLock.RLock()
	defer h.ClientLock.RUnlock()
	if c, ok := h.Cid2Client[clientId]; ok {
		select {
		case c.Msgs <- resp:
			logger.Infof("[stream] SendByClientId success clientId:%s resp:%v", clientId, string(bb))
			return nil
		default:
			logger.Infof("[stream] SendByClientId success clientId:%s resp:%v queue max", clientId, string(bb))
			c.Msgs <- resp
		}
	}
	logger.Infof("[stream] SendByClientId fail clientId:%s resp:%v 客户端不在线", clientId, string(bb))
	return errors.New("客户端不在线")
}

// 广播订阅代码
func (h *Hub) broadcast(changeCode string) {
	// 获取分类
	h.CategoryLock.RLock()
	registerInfo, registerOk := h.getCategoryByCode(changeCode)
	h.CategoryLock.RUnlock()
	if !registerOk {
		return
	}

	// 获取客户端
	h.ClientLock.RLock()
	clientIds := make([]string, 0)
	for _, v := range h.Cid2Client {
		if v.ClientInfo.SubscribeCode == changeCode {
			clientIds = append(clientIds, v.ClientInfo.ClientId)
		}
	}
	h.ClientLock.RUnlock()

	// 广播订阅代码的客户端
	registerInfo.Lock.RLock()
	sendMsg := registerInfo.packMsg()
	registerInfo.Lock.RUnlock()

	for _, cid := range clientIds {
		h.sendByClientId(cid, sendMsg)
	}
}

// 相关分类缓存管理
func (h *Hub) loadFromDb() error {
	list, err := model.CategoryList(nil)
	if err != nil {
		return err
	}

	hh, _ := json.Marshal(list)
	logger.Infof("hh:%s", string(hh))

	h.CategoryLock.Lock()
	defer h.CategoryLock.Unlock()

	for _, v := range list {
		_, err := h.addCategoryByModel(v)
		if err != nil {
			panic(err)
		}
	}
	return nil
}

func (h *Hub) removeNodeAndBroadcast(clientInfo *ClientInfo) {
	h.CategoryLock.RLock()
	registerInfo, registerOk := h.getCategoryByCode(clientInfo.RegisterCode)
	h.CategoryLock.RUnlock()

	if !registerOk {
		return
	}

	registerInfo.Lock.Lock()
	oldVersion := registerInfo.LastVersion

	if clientInfo.RegisterCode == define.CategoryRoute {
		// 路由节点无需广播
		registerInfo.removeNode(clientInfo.ClientId)
	} else {
		//绑定节点
		registerInfo.unbingNode(clientInfo.ClientId)
	}
	registerInfo.Lock.Unlock()

	if oldVersion != registerInfo.LastVersion {
		h.broadcast(clientInfo.RegisterCode)
	}
	logger.Infof("[stream] removeNodeAndBroadcast clientId:%s", clientInfo.ClientId)
}

func (h *Hub) getCategoryByCode(code string) (*CategoryInfo, bool) {
	v, ok := h.CategoryMap[code]
	return v, ok
}

func (h *Hub) addCategoryByModel(m *model.CategoryModel) (*CategoryInfo, error) {
	categoryInfo := &CategoryInfo{
		List:        make([]*pb.CategoryNodeItem, 0),
		Map:         make(map[string]*pb.CategoryNodeItem),
		LastVersion: 0,
		Lock:        new(sync.RWMutex),
		Code:        m.Code,
	}
	contentList, err := category.IsContentValid(m.Content)
	if err != nil {
		return nil, err
	}

	for _, v2 := range contentList {
		item := &pb.CategoryNodeItem{
			Index:  v2.Index,
			Start:  v2.Start,
			End:    v2.End,
			Ip:     "",
			Port:   "",
			Status: define.NodeUnBind,
		}
		categoryInfo.List = append(categoryInfo.List, item)
	}

	h.addCategory(categoryInfo)
	return categoryInfo, nil
}

func (h *Hub) addCategory(c *CategoryInfo) {
	_, ok := h.CategoryMap[c.Code]
	if ok {
		panic(fmt.Sprintf("[stream] addCategory code:%s exist", c.Code))
	}
	h.CategoryMap[c.Code] = c
	bb, _ := json.Marshal(c)
	logger.Infof("[stream] addCategory 添加分类 code:%s data:%s", c.Code, string(bb))
}

func (h *Hub) reloadCategory(categoryModel *model.CategoryModel) error {
	contentList, err := category.IsContentValid(categoryModel.Content)
	if err != nil {
		return err
	}

	if categoryModel.IsDelete == 1 {
		contentList = make([]*pb.CategoryNodeItem, 0)
	}
	h.CategoryLock.Lock()
	categoryInfo, ok := h.CategoryMap[categoryModel.Code]
	oldData := []byte("[]")
	// 添加配置
	if !ok {
		categoryInfo, err = h.addCategoryByModel(categoryModel)
		if err != nil {
			logger.Infof("[stream] addCategoryByModel fail err:%s", err.Error())
		}
		h.CategoryLock.Unlock()
	} else {
		oldData, _ = json.Marshal(categoryInfo)
		h.CategoryLock.Unlock()

		categoryInfo.Lock.Lock()
		// 内容更新 删除 增加 修改
		oldMap := make(map[string]*pb.CategoryNodeItem)
		newMap := make(map[string]*pb.CategoryNodeItem)
		for _, v2 := range contentList {
			newMap[v2.Index] = v2
		}
		for _, v2 := range categoryInfo.List {
			oldMap[v2.Index] = v2
		}

		oldVersion := categoryInfo.LastVersion
		// 删除
		removeList := make([]*pb.CategoryNodeItem, 0)
		for index, oldInfo := range oldMap {
			if _, ok := newMap[index]; !ok {
				delete(oldMap, index)
				for i, v2 := range categoryInfo.List {
					if v2.Index == index {
						categoryInfo.List = append(categoryInfo.List[:i], categoryInfo.List[i+1:]...)
						break
					}
				}
				removeList = append(removeList, oldInfo)
				if oldInfo.Status == define.NodeBind {
					h._removeByClientId(GetClientId(oldInfo.Ip, oldInfo.Index))
				}
				categoryInfo.LastVersion++

				logger.Infof("[stream] reloadCategory 删除节点 code:%s index:%s data:%v", categoryModel.Code, oldInfo.Index, oldInfo)
			}
		}
		// 增加配置节点
		for index, v := range newMap {
			if _, ok := oldMap[index]; !ok {
				categoryInfo.List = append(categoryInfo.List, v)
				categoryInfo.LastVersion++
				logger.Infof("[stream] reloadCategory 添加节点 code:%s index:%s data:%v", categoryModel.Code, v.Index, v)
			}
		}
		// 更新
		for index, v := range newMap {
			if v2, ok := oldMap[index]; ok {
				v2.Start = v.Start
				v2.End = v.End
				categoryInfo.LastVersion++
				logger.Infof("[stream] reloadCategory 更新节点 code:%s index:%s data:%v", categoryModel.Code, v.Index, v2)
			}
		}
		categoryInfo.Lock.Unlock()
		if oldVersion != categoryInfo.LastVersion {
			h.broadcast(categoryInfo.Code)
		}

	}
	newData, _ := json.Marshal(categoryInfo)
	logger.Infof("[stream] reloadCategory code:%s 更新前数据:%s 更新后数据:%s", categoryModel.Code, string(oldData), string(newData))

	return nil
}
