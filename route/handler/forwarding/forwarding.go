package forwarding

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/visonlv/go-vkit/errorsx"
	"github.com/visonlv/go-vkit/grpcclient"
	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/go-vkit/utilsx"
	"github.com/visonlv/iot-engine/common/define"
	"github.com/visonlv/iot-engine/common/utils"
	grouppb "github.com/visonlv/iot-engine/group/proto"
	pb "github.com/visonlv/iot-engine/shadow/proto"
	shadowpb "github.com/visonlv/iot-engine/shadow/proto"
)

type GrpcClient struct {
	addr string
	c    *pb.ForwardingServiceClient
	node *grouppb.CategoryNodeItem
}

type ClientReq struct {
	c   *pb.ForwardingServiceClient
	sns []string
}

type PropertiesResult struct {
	resp *pb.ForwardingPropertiesResp
	err  error
}

type WatchResult struct {
	resp *pb.ForwardingWatchResp
	err  error
}

type WatchContext struct {
	cancel   context.CancelFunc
	groupMap map[int32]int32
}

type Forwarding struct {
	//所有客户端
	clients map[string]*GrpcClient
	//分组数据->客户端关系
	group2client map[int32]*GrpcClient
	//最新版本号
	LastVersion int64
	//节点信息
	addr2GroupItemMap map[string]*grouppb.CategoryNodeItem
	//断开存在client不对应的watch上下文
	contextId2Context map[string]*WatchContext
	reloadLock        *sync.RWMutex
}

func newForwarding() *Forwarding {
	return &Forwarding{
		clients:           make(map[string]*GrpcClient),
		group2client:      make(map[int32]*GrpcClient),
		LastVersion:       -1,
		addr2GroupItemMap: make(map[string]*grouppb.CategoryNodeItem),
		contextId2Context: make(map[string]*WatchContext),
		reloadLock:        new(sync.RWMutex),
	}
}

func (f *Forwarding) reloadClient(data *grouppb.CategoryHeartBeatResp) error {
	if data.Code != 0 {
		logger.Infof("ReloadClient fail code:%d msg:%s", data.Code, data.Msg)
		return fmt.Errorf("code:%d msg:%s", data.Code, data.Msg)
	}

	// 先判断版本号跟数据是否一致
	f.reloadLock.RLock()
	if data.LastVersion == f.LastVersion {
		logger.Infof("not need reload data:%v", data)
		f.reloadLock.RUnlock()
		return nil
	}
	f.reloadLock.RUnlock()

	addr2ItemMap := make(map[string]*grouppb.CategoryNodeItem)
	group2Addr := make(map[int32]string)
	for _, v := range data.Items {
		if v.Ip == "" {
			continue
		}
		key := getItemKey(v)
		addr2ItemMap[key] = v
		for i := v.Start; i <= v.End; i++ {
			group2Addr[i] = key
		}
	}
	f.reloadLock.Lock()
	defer f.reloadLock.Unlock()
	//移掉不存在的client
	for k, _ := range f.clients {
		if _, ok := addr2ItemMap[k]; !ok {
			grpcclient.DelConnClient(k)
			delete(f.clients, k)
			logger.Infof("delete client:%s", k)
		}
	}
	//移掉不存在的client对应的group
	deleteGroup := make(map[int32]int32)
	for group, v := range f.group2client {
		addr, ok := group2Addr[group]
		if !ok || v.addr != addr {
			delete(f.group2client, group)
			deleteGroup[group] = group
			logger.Infof("delete group:%d", group)
		}
	}
	//添加新增的client
	for k, v := range addr2ItemMap {
		if _, ok := f.clients[k]; !ok {
			conn := grpcclient.GetConnClient(
				k,
				grpcclient.RequestTimeout(time.Second*20),
			)
			service := shadowpb.NewForwardingServiceClient("shadow", conn)
			f.clients[k] = &GrpcClient{
				addr: k,
				c:    service,
				node: v,
			}
			logger.Infof("add client:%s", k)
		}
	}
	//添加新增的client对应的group
	for group, addr := range group2Addr {
		if _, ok := f.group2client[group]; !ok {
			c, ok1 := f.clients[addr]
			if !ok1 {
				panic(fmt.Sprintf("client:%s not exist", addr))
			}
			f.group2client[group] = c
			logger.Infof("add group:%d", group)
		}
	}

	// 移除上下文
	for _, group := range deleteGroup {
		for id, v := range f.contextId2Context {
			if _, ok := v.groupMap[group]; ok {
				delete(f.contextId2Context, id)
				v.cancel()
			}
		}
	}

	//检查是否所有group都有关系
	groupCount := 0
	for group, _ := range group2Addr {
		if group < define.MinGroup || group > define.MaxGroup {
			panic(fmt.Sprintf("group:%d in(0-99)", group))
		}
		groupCount++
	}

	//修改版本号
	f.LastVersion = data.LastVersion
	//修改数据
	f.addr2GroupItemMap = addr2ItemMap
	return nil
}

func (f *Forwarding) getClient(sn string) (*shadowpb.ForwardingServiceClient, error) {
	f.reloadLock.RLock()
	defer f.reloadLock.RUnlock()

	groupIndex := utils.GetGroupId(sn)
	c, ok := f.group2client[groupIndex]
	if !ok {
		return nil, fmt.Errorf("group:%d not found", groupIndex)
	}
	return c.c, nil
}

func (f *Forwarding) getNeedReqClients(pks, sns []string) (map[string]*ClientReq, error) {
	clientMap := make(map[string]*ClientReq, 0)
	f.reloadLock.RLock()
	if len(sns) > 0 {
		for _, sn := range sns {
			groupIndex := utils.GetGroupId(sn)
			c, ok := f.group2client[groupIndex]
			if !ok {
				return nil, fmt.Errorf("group:%d not found", groupIndex)
			}

			clientReq, ok := clientMap[c.addr]
			if !ok {
				clientReq = &ClientReq{
					c:   c.c,
					sns: make([]string, 0),
				}
			}
			clientReq.sns = append(clientReq.sns, sn)
			clientMap[c.addr] = clientReq
		}
	} else {
		for _, v := range f.clients {
			clientReq := &ClientReq{
				c:   v.c,
				sns: pks,
			}
			clientMap[v.addr] = clientReq
		}
	}
	f.reloadLock.RUnlock()

	return clientMap, nil
}

func (f *Forwarding) properties(ctx context.Context, req *shadowpb.ForwardingPropertiesReq, resp *shadowpb.ForwardingPropertiesResp) error {
	clientMap, err := f.getNeedReqClients(req.Pks, req.Sns)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = err.Error()
		return nil
	}
	clientLength := len(clientMap)
	if clientLength <= 0 {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = "not any target for call"
		return nil
	}

	resultCh := make(chan *PropertiesResult)
	newCtx, cancel := context.WithCancel(ctx)
	for _, v := range clientMap {
		service := v.c
		sns := v.sns
		go func() {
			newResp, err := service.Properties(newCtx, &shadowpb.ForwardingPropertiesReq{
				Pks:   req.Pks,
				Sns:   sns,
				Codes: req.Codes,
			})

			bb, _ := json.Marshal(newResp)
			logger.Infof("newResp:%s", string(bb))
			select {
			case <-newCtx.Done():
				return
			default:
				resultCh <- &PropertiesResult{newResp, err}
			}
		}()
	}

	resp.List = make([]*shadowpb.ForwardingProperty, 0)
	for i := 0; i < clientLength; i++ {
		r := <-resultCh
		if r.err != nil {
			cancel()
			return r.err
		}

		if r.resp.Code != 0 {
			cancel()
			resp.Code = r.resp.Code
			resp.Msg = r.resp.Msg
			return nil
		}
		resp.List = append(resp.List, r.resp.List...)
	}
	cancel()
	return nil
}

func (f *Forwarding) watch(ctx context.Context, req *shadowpb.ForwardingWatchReq, s *shadowpb.ForwardingService_WatchServer) error {
	defer func() {
		logger.Infof("[forwarding] watch contextId:%s finish", req.ContextId)
	}()

	clientMap, err := f.getNeedReqClients(req.Pks, req.Sns)
	if err != nil {
		err := s.Send(&shadowpb.ForwardingWatchResp{
			Code: errorsx.FAIL.Code,
			Msg:  err.Error(),
		})
		if err != nil {
			logger.Infof("[forwarding] watch contextId:%s send fail:%s", req.ContextId, err.Error())
		}
		return nil
	}
	clientLength := len(clientMap)
	if clientLength <= 0 {
		err := s.Send(&shadowpb.ForwardingWatchResp{
			Code: errorsx.FAIL.Code,
			Msg:  "not any target for call",
		})
		if err != nil {
			logger.Infof("[forwarding] watch contextId:%s send fail2:%s", req.ContextId, err.Error())
		}
		return nil
	}

	resultCh := make(chan *WatchResult, 1024)
	ctxAll, cancelAll := context.WithCancel(ctx)
	groupMap := make(map[int32]int32)
	for _, v := range clientMap {
		service := v.c
		sns := v.sns
		for _, sn := range sns {
			groupIndex := utils.GetGroupId(sn)
			groupMap[groupIndex] = groupIndex
		}
		go func() {
			ctxOne, cancelOne := context.WithCancel(ctxAll)
			child, err := service.Watch(ctxOne, &shadowpb.ForwardingWatchReq{
				Pks:      req.Pks,
				Sns:      sns,
				MsgTypes: req.MsgTypes,
				Codes:    req.Codes,
			})
			defer func() {
				logger.Infof("[forwarding] watch contextId:%s send fail2:%s", req.ContextId, err.Error())
				cancelOne()
			}()

			if err != nil {
				select {
				case <-ctxAll.Done():
					logger.Infof("[forwarding] watch contextId:%s ctxAll.Done() err:%s", req.ContextId, err.Error())
					return
				case <-ctxOne.Done():
					logger.Infof("[forwarding] watch contextId:%s ctxOne.Done() err:%s", req.ContextId, err.Error())
					return
				default:
					logger.Infof("[forwarding] watch contextId:%s err:%s", req.ContextId, err.Error())
					resultCh <- &WatchResult{err: err}
					return
				}
			}

			for {
				select {
				case <-ctxAll.Done():
					logger.Infof("[forwarding] watch contextId:%s for ctxAll.Done()", req.ContextId)
					return
				case <-ctxOne.Done():
					logger.Infof("[forwarding] watch contextId:%s for ctxOne.Done()", req.ContextId)
					return
				default:
					msg, err := child.Recv()
					select {
					case <-ctxAll.Done():
						logger.Infof("[forwarding] watch contextId:%s read ctxAll.Done()", req.ContextId)
						return
					case <-ctxOne.Done():
						logger.Infof("[forwarding] watch contextId:%s read ctxOne.Done()", req.ContextId)
						return
					default:
						if err != nil {
							logger.Infof("[forwarding] watch contextId:%s read default err:%s", req.ContextId, err.Error())
							resultCh <- &WatchResult{err: err}
							return
						}
						if err == nil {
							logger.Infof("[forwarding] watch contextId:%s read default success", req.ContextId)
							resultCh <- &WatchResult{err: nil, resp: msg}
						}
					}
				}
			}
		}()
	}

	contextId := utilsx.GenUuid()
	f.contextId2Context[contextId] = &WatchContext{
		cancel:   cancelAll,
		groupMap: groupMap,
	}
	defer func() {
		delete(f.contextId2Context, contextId)
	}()
	for {
		select {
		case <-ctxAll.Done():
			cancelAll()
			return nil
		case watchResult := <-resultCh:
			//出现异常统一取消
			if watchResult.err != nil {
				logger.Infof("[forwarding] watch contextId:%s fail:%s", req.ContextId, watchResult.err.Error())
				cancelAll()
				return watchResult.err
			}
			err := s.Send(watchResult.resp)
			if err != nil {
				logger.Infof("[forwarding] watch contextId:%s send fail:%s", req.ContextId, watchResult.err.Error())
				cancelAll()
				return nil
			}
		}
	}
}
