package forwarding

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/go-vkit/utilsx"
	commonclient "github.com/visonlv/iot-engine/common/client"
	commonnats "github.com/visonlv/iot-engine/common/client/nats"
	"github.com/visonlv/iot-engine/common/define"
	"github.com/visonlv/iot-engine/common/proto/messaging"
	"github.com/visonlv/iot-engine/common/utils"
	"github.com/visonlv/iot-engine/shadow/app"
	"github.com/visonlv/iot-engine/shadow/model"
	pb "github.com/visonlv/iot-engine/shadow/proto"
	thingpb "github.com/visonlv/iot-engine/thing/proto"
	"google.golang.org/protobuf/proto"
)

const (
	notUseEventType             int = 0
	addOrUpdateProductEventType int = iota
	deleteProductEventType
	natsEventType
	addWatchEventType
	cancelWatchEventType
	asynSendMsgEventType
	synSendMsgEventType
	cancelContextEventType
	saveShadowResultEventType
	getDeviceListEventType
	getProductMapEventType
)

type ChanEvent struct {
	eventType int
	param     interface{}
}

type ChanInEvent struct {
	outCh chan *ChanEvent
	in    *ChanEvent
}

type synSendMsgEventReq struct {
	msg *SendMsgReq
	ctx context.Context
}

type synSendMsgEventResp struct {
	msg *SendMsgResp
	err error
}

type addWatchEventReq struct {
	contextId      string
	ctx            context.Context
	watchSourceReq *pb.ForwardingWatchReq
}

type getProductMapResp struct {
	pk2Product map[string]*Product
	err        error
}

type executorChild struct {
	ExecutorChildIndex int
	//nats代理 发送消息到nats
	natsClient *commonnats.NatsClient
	//所有产品内存维护，只是内存地址多分复制
	products *Products
	//设备数据缓存 使用lru控制缓存有效性
	devices *Devices
	//脏数据额外标识，避免全局lru遍历
	dirtyShadows *DirtyShadows
	//下行消息上下文
	downMsgContexts *DownMsgContexts
	//统一消息入口
	inChan chan *ChanInEvent
	//订阅关系
	match *Match
	//订阅上下文
	contextId2SubList map[string][]*subscription
	//消息处理器
	eventHandler *EventHandler
	//物模型在保存中
	shadowSaving     int32
	shadowSavingLock *sync.Mutex
}

func newExecutorChild(ExecutorChildIndex int) *executorChild {
	products := &Products{pk2Product: make(map[string]*Product)}
	dirtyShadows := &DirtyShadows{delayEvents: make(map[string]*DelayEvent)}
	downMsgContexts := &DownMsgContexts{id2Context: make(map[string]*DownMsgContext)}
	s := &executorChild{
		ExecutorChildIndex: ExecutorChildIndex,
		products:           products,
		devices:            newDevices(10, products),
		dirtyShadows:       dirtyShadows,
		downMsgContexts:    downMsgContexts,
		inChan:             make(chan *ChanInEvent, 128),
		match:              newMatch(),
		contextId2SubList:  make(map[string][]*subscription),
		eventHandler:       newEventHandler(),
		shadowSaving:       0,
		shadowSavingLock:   new(sync.Mutex),
	}
	return s
}

func (s *executorChild) handleMsg(event *ChanInEvent) {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("[forwarding]: handleMsg %s", string(debug.Stack()))
			logger.Errorf("[forwarding]: handleMsg err:%s", err)
		}
	}()

	eventIn := event.in
	eventType := eventIn.eventType
	eventParam := eventIn.param
	outCh := event.outCh
	switch eventType {
	case addOrUpdateProductEventType:
		s.products.AddOrUpdateProduct(eventParam.(*Product))
	case deleteProductEventType:
		s.products.DelProduct(eventParam.(*Product))
	case natsEventType:
		s.handlerNatsMsg(eventParam.(*messaging.Message))
	case addWatchEventType:
		s.addWatch(eventParam.(*addWatchEventReq), outCh)
	case cancelWatchEventType:
		s.cancelWatch(eventParam.(string))
	case asynSendMsgEventType:
		err := s.asynSendMsg(eventParam.(*SendMsgReq))
		outCh <- &ChanEvent{eventType: eventIn.eventType, param: err}
	case synSendMsgEventType:
		err := s.syncSendMsg(eventParam.(*synSendMsgEventReq), outCh)
		if err != nil {
			outCh <- &ChanEvent{
				eventType: eventIn.eventType,
				param:     &synSendMsgEventResp{err: err},
			}
		}
	case cancelContextEventType:
		s.cancelContext(eventParam.(string))
	case saveShadowResultEventType:
		s.saveShadowResult(eventParam.([]*saveShadowItem))
	case getDeviceListEventType:
		outCh <- &ChanEvent{
			eventType: eventIn.eventType,
			param:     s.getDeviceList(eventParam.([]string)),
		}
	case getProductMapEventType:
		outCh <- &ChanEvent{
			eventType: eventIn.eventType,
			param:     s.getProductMap(eventParam.([]string)),
		}
	default:
		logger.Infof("[forwarding] not suport eventType:%d", eventType)
		return
	}
}

func (s *executorChild) start(list []*Product) error {
	for _, v := range list {
		s.products.AddOrUpdateProduct(v)
	}
	//nats 启动
	natsClient, err := commonnats.NewClient(app.Cfg.Nats.Url, app.Cfg.Nats.Username, app.Cfg.Nats.Password, fmt.Sprintf("executor_child:%d", s.ExecutorChildIndex), "")
	if err != nil {
		logger.Errorf("[forwarding] Nats2Local failed to connect to nats broker: %s", err)
		return err
	}
	s.natsClient = natsClient
	s.subNatsSysEvent()

	go s.mainloop()

	return nil
}

func pbProduct2Propruct(data []byte, withThingInfo bool) (*Product, error) {
	p := &thingpb.Product{}
	err := proto.Unmarshal(data, p)
	if err != nil {
		logger.Infof("[forwarding] pbProduct2Propruct decode pb err:%s", err.Error())
		return nil, fmt.Errorf("pbProduct2Propruct decode pb err:%s", err.Error())
	}

	m, err := convertProductPb2Info(p)
	if err != nil {
		return nil, fmt.Errorf("pbProduct2Propruct convertProductPb2Info err:%s", err.Error())
	}

	return m, nil
}

func (s *executorChild) subNatsSysEvent() {
	s.natsClient.SubscribeNatsMsg(context.Background(), "executor_child", define.SysTopicProductUpdate, func(msg *nats.Msg) {
		p, err := pbProduct2Propruct(msg.Data, true)
		if err != nil {
			logger.Infof("[forwarding] define.SysTopicProductUpdate msg err:%s", err.Error())
			return
		}
		s.ayncAddSimpleEvent(addOrUpdateProductEventType, p)
	})

	s.natsClient.SubscribeNatsMsg(context.Background(), "executor_child", define.SysTopicProductAdd, func(msg *nats.Msg) {
		p, err := pbProduct2Propruct(msg.Data, true)
		if err != nil {
			logger.Infof("[forwarding] define.SysTopicProductAdd msg err:%s", err.Error())
			return
		}
		s.ayncAddSimpleEvent(addOrUpdateProductEventType, p)
	})

	s.natsClient.SubscribeNatsMsg(context.Background(), "executor_child", define.SysTopicProductDel, func(msg *nats.Msg) {
		p, err := pbProduct2Propruct(msg.Data, false)
		if err != nil {
			logger.Infof("[forwarding] define.SysTopicProductDel msg err:%s", err.Error())
			return
		}
		s.ayncAddSimpleEvent(deleteProductEventType, p)
	})
}

func (s *executorChild) mainloop() error {
	t := time.NewTicker(time.Second)
	defer t.Stop()

	for {
		select {
		case msg := <-s.inChan:
			s.handleMsg(msg)
		case <-t.C:
			s.batchSaveShadow()
		}
	}
}

func (s *executorChild) batchSaveShadow() {
	s.shadowSavingLock.Lock()
	if s.shadowSaving == 1 {
		s.shadowSavingLock.Unlock()
		return
	}
	s.shadowSaving = 1
	s.shadowSavingLock.Unlock()

	dirtyList := make([]*saveShadowItem, 0)
	s.devices.deviceLru.RangeDirty(func(key string, value *Device) bool {
		bb, _ := json.Marshal(value.Shadow)
		dirtyList = append(dirtyList, &saveShadowItem{
			sn:      key,
			shadow:  string(bb),
			version: value.ShadowVersion,
		})
		logger.Infof("[forwarding] batchSaveShadow sn:%s version:%d", key, value.ShadowVersion)
		return true
	})

	go func() {
		defer func() {
			s.shadowSavingLock.Lock()
			s.shadowSaving = 0
			s.shadowSavingLock.Unlock()

			s.ayncAddSimpleEvent(saveShadowResultEventType, dirtyList)
		}()

		for _, v := range dirtyList {
			err := model.ShadowUpdateBySnAndShadowAndVersion(nil, v.sn, v.shadow, v.version)
			if err != nil {
				logger.Errorf("[forwarding] ShadowUpdateBySnAndShadowAndVersion sn:%s shadow:%s version:%d fail:%s", v.sn, v.shadow, v.version, err.Error())
			} else {
				logger.Infof("[forwarding] ShadowUpdateBySnAndShadowAndVersion sn:%s shadow:%s version:%d success", v.sn, v.shadow, v.version)
			}
		}
	}()
}

func (s *executorChild) saveShadowResult(dirtyList []*saveShadowItem) {
	for _, v := range dirtyList {
		if s.devices.deviceLru.CleanDirty(v.sn, v.version) {
			logger.Infof("[forwarding] saveShadowResult CleanDirty sn:%s shadow:%s version:%d success", v.sn, v.shadow, v.version)
		}
	}
}

func (s *executorChild) getDeviceList(deviceSn []string) []*Device {
	devices := make([]*Device, 0)
	if len(deviceSn) == 0 {
		s.devices.deviceLru.Range(func(key string, value *Device) bool {
			devices = append(devices, value)
			return true
		})
	} else {
		for _, v := range deviceSn {
			d, ok := s.devices.deviceLru.Get(v)
			if ok {
				devices = append(devices, d)
			}
		}
	}
	return devices
}

func (s *executorChild) getProductMap(pks []string) *getProductMapResp {
	pk2Product := make(map[string]*Product, 0)
	for _, v := range pks {
		p, err := s.products.GetProduct(v)
		if err != nil {
			return &getProductMapResp{err: err}
		}
		pk2Product[v] = p
	}
	return &getProductMapResp{pk2Product: pk2Product}
}

func (s *executorChild) asyncAddEvent(event *ChanInEvent) {
	select {
	case s.inChan <- event:
		return
	default:
		logger.Infof("[forwarding] ExecutorChildIndex:%d asyncAddEvent blook", s.ExecutorChildIndex)
		s.inChan <- event
	}
}

func (s *executorChild) ayncAddSimpleEvent(eventType int, param interface{}) {
	event := &ChanInEvent{
		in: &ChanEvent{eventType: eventType, param: param},
	}
	s.asyncAddEvent(event)
}

func (s *executorChild) asynSendMsg(msg *SendMsgReq) error {
	dInfo, pInfo, globalErr := s.devices.getDeviceAndProduct(msg.Sn)
	if globalErr != nil {
		return globalErr
	}

	if dInfo.PSn == msg.Sn {
		logger.Infof("[forwarding] asynSendMsg 网关sn跟子设备sn不能一样 sn:%s msg:%s", msg.Sn, utils.JsonToString(msg))
		return fmt.Errorf("网关sn跟子设备sn一样")
	}

	directPublish := dInfo.PSn == ""
	var newMsg *SendMsgReq
	// 需要通过网关转发
	if !directPublish {
		dInfoGateway, _, err := s.devices.getDeviceAndProduct(dInfo.PSn)
		if err != nil {
			return err
		}

		if !dInfoGateway.Shadow.PropertyEqualTo(define.PropertyOnline, true) {
			globalErr = fmt.Errorf("网关不在线")
			goto saveMsg
		}

		// 重新构造网关数据
		sourcePayload := make(map[string]any)
		err = json.Unmarshal(msg.Payload, &sourcePayload)
		if err != nil {
			globalErr = fmt.Errorf("不是标准json格式:%s", err.Error())
			goto saveMsg
		}

		sourcePayload["identify"] = msg.Code
		newPayload := &define.GateWayDownPayload{}
		newPayload.Id = utilsx.GenUuid()
		newPayload.ContextId = utilsx.GenUuid()
		newPayload.Time = time.Now().UnixMilli()
		newPayload.SubMsgs = make([]*define.GateWayDownChileMsg, 0)
		newPayload.SubMsgs = append(newPayload.SubMsgs, &define.GateWayDownChileMsg{
			Sn:  msg.Sn,
			Msg: sourcePayload,
		})
		newPayloadbb, _ := json.Marshal(newPayload)
		newMsg = &SendMsgReq{
			Sn:        dInfo.PSn,
			ContextId: utilsx.GenUuid(),
			Code:      define.MsgCodeGatewayProxy,
			MsgType:   msg.MsgType,
			Payload:   newPayloadbb,
			Timeout:   msg.Timeout,
		}

	}

	//1、判断设备是否在线
	if globalErr == nil && !dInfo.Shadow.PropertyEqualTo(define.PropertyOnline, true) {
		globalErr = fmt.Errorf("设备不在线")
		goto saveMsg
	}

	// 消息格式
saveMsg:
	topic := fmt.Sprintf("iot.down.%d.%s.%s.%s.%s", dInfo.Group, pInfo.Pk, dInfo.Sn, msg.MsgType, msg.Code)
	m := &messaging.Message{
		Id:        utilsx.GenUuid(),
		ContextId: msg.ContextId,
		Pk:        pInfo.Pk,
		Sn:        dInfo.Sn,
		Topic:     topic,
		Transform: string(pInfo.Transform),
		Protocol:  string(pInfo.Protocol),
		Supplier:  "iot-engine",
		Payload:   msg.Payload,
		Created:   time.Now().UnixMilli(),
	}

	if globalErr == nil && !directPublish {
		_, globalErr = s.eventHandler.decodeDownMsg(msg.MsgType, msg.Code, m, pInfo, dInfo)
		if globalErr != nil {
			globalErr = fmt.Errorf("[forwarding] decodeDownMsg decode msg fail:%s", globalErr.Error())
		}
	}

	if globalErr == nil && directPublish {
		globalErr = s.natsClient.Publish(context.Background(), topic, m)
	}

	result := "{}"
	if globalErr != nil {
		result = fmt.Sprintf("{\"code\":-1,\"msg\":\"%s\"}", globalErr.Error())
	}

	err1 := model.MsgLogInsert(context.Background(), &model.MsgLog{
		Pk:        m.Pk,
		Sn:        m.Sn,
		Content:   string(m.Payload),
		Topic:     strings.ReplaceAll(m.Topic, ".", "/"),
		LogType:   msg.MsgType,
		Ts:        time.Now(),
		MsgId:     m.Id,
		ContextId: m.ContextId,
		Result:    result,
		Dir:       string(define.ServiceDirDown),
		Code:      msg.Code,
	})

	if globalErr != nil {
		return globalErr
	}

	if err1 != nil {
		return err1
	}

	if globalErr == nil && !directPublish {
		return s.asynSendMsg(newMsg)
	}

	return nil
}

func (s *executorChild) syncSendMsg(req *synSendMsgEventReq, outCh chan *ChanEvent) error {
	msg := req.msg
	err := s.asynSendMsg(msg)
	if err != nil {
		return err
	}
	// 添加上下文
	return s.downMsgContexts.Add(&DownMsgContext{
		contextId: msg.ContextId,
		ctx:       req.ctx,
		outCh:     outCh,
	})
}

func (s *executorChild) cancelContext(contextId string) {
	s.downMsgContexts.Del(contextId)
}

// 统一处理nats消息
func (s *executorChild) handlerNatsMsg(msg *messaging.Message) {
	pk, sn, group, msgType, code, isUp, err := commonclient.DecodeNatsTopic(msg.Topic)
	if err != nil {
		logger.Infof("[forwarding] handlerNatsMsg decode topic fail:%s err:%s", msg.Topic, err.Error())
		return
	}

	logger.Infof("[forwarding] handlerNatsMsg topic:%s decode result pk:%v sn:%v group:%v msgType:%v code:%v isUp:%v", msg.Topic, pk, sn, group, msgType, code, isUp)

	dInfo, product, err := s.devices.getDeviceAndProduct(msg.Sn)
	if err != nil {
		logger.Infof("[forwarding] handlerNatsMsg getDevice fail:%s err:%s", msg.Topic, err.Error())
		return
	}

	// 网关代理信息特殊处理
	if code == define.MsgCodeGatewayProxy {
		decodeResult, err := s.eventHandler.decodeGatewayUpMsg(msgType, code, msg)
		if err != nil {
			logger.Errorf("[forwarding] handlerNatsMsg decode msg fail:%s", err.Error())
			return
		}
		//批量转发
		s.tryPublishChildMsg(decodeResult)
		s.trySaveLog(decodeResult, code)
		return
	}

	decodeResult, err := s.eventHandler.decodeUpMsg(msgType, code, msg, product, dInfo)
	if err != nil {
		logger.Errorf("[forwarding] handlerNatsMsg decode msg fail:%s", err.Error())
		return
	}
	s.trySaveShadow(decodeResult)
	s.tryWaitUpContext(decodeResult)
	s.tryHandlerWatchEvent(decodeResult)
	s.trySaveLog(decodeResult, code)
}

func (s *executorChild) trySaveLog(result *DecodeResult, code string) {
	err := model.MsgLogInsert(context.Background(), &model.MsgLog{
		Pk:        result.NatsMsg.Pk,
		Sn:        result.NatsMsg.Sn,
		Content:   string(result.NatsMsg.Payload),
		Topic:     strings.ReplaceAll(result.NatsMsg.Topic, ".", "/"),
		LogType:   result.MsgType,
		Dir:       string(define.ServiceDirUp),
		Ts:        time.Now(),
		MsgId:     result.PayloadCommonResult.Id,
		ContextId: result.PayloadCommonResult.ContextId,
		Code:      code,
		Result:    "{}",
	})

	if err != nil {
		logger.Infof("insert fail err:%s", err)
	}
}

func (s *executorChild) trySaveShadow(result *DecodeResult) {
	msgType := result.MsgType
	if msgType == define.MsgTypeProperty {
		hasChange := false
		// 2、属性存库
		info := result.PayloadResult.(*define.UpPropertyPayload)
		for k, v := range info.Params {
			propertyInfo, ok := result.Device.Shadow.Properties[k]
			if !ok {
				propertyInfo = &define.ShadowProperty{
					Current: &define.CurrentProperty{},
				}
			}
			propertyInfo.Current.Value = v
			propertyInfo.Current.UpdatedTime = time.Now().UnixMilli()
			result.Device.Shadow.Properties[k] = propertyInfo
			hasChange = true
		}
		if hasChange {
			result.Device.ShadowVersion++
			s.devices.deviceLru.Set(result.NatsMsg.Sn, result.Device, result.Device.ShadowVersion)
		}
	}
}

func (s *executorChild) tryWaitUpContext(result *DecodeResult) {
	msgType := result.MsgType
	msg := result.NatsMsg
	if msgType == define.MsgTypePropertyReply || msgType == define.MsgTypeServiceReply {
		if msg.ContextId != "" {
			downContext, err := s.downMsgContexts.GetAndDel(msg.ContextId)
			if err != nil {
				logger.Infof("[forwarding] downContext index:%d contextId:%s err:%s", s.ExecutorChildIndex, msg.ContextId, err.Error())
				return
			}
			select {
			case <-downContext.ctx.Done():
				logger.Infof("[forwarding] downContext index:%d contextId:%s isDone", s.ExecutorChildIndex, msg.ContextId)
				return
			default:
				resp := &synSendMsgEventResp{msg: &SendMsgResp{M: msg}}
				downContext.outCh <- &ChanEvent{eventType: synSendMsgEventType, param: resp}
				logger.Infof("[forwarding] downContext index:%d contextId:%s sended", s.ExecutorChildIndex, msg.ContextId)
				return
			}
		}
	}
}

func (s *executorChild) tryHandlerWatchEvent(result *DecodeResult) {
	msgType := result.MsgType
	msg := result.NatsMsg
	pk, sn, _, msgType, code, _, err := commonclient.DecodeNatsTopic(msg.Topic)
	if err != nil {
		logger.Errorf("[forwarding] tryHandlerWatchEvent index:%d DecodeNatsTopic topic:%s fail:%s ", s.ExecutorChildIndex, msg.Topic, err.Error())
		return
	}
	tokens := []string{pk, sn, msgType, code}
	r := s.match.Match(tokens)
	if len(r.psubs) == 0 {
		return
	}
	e := &ChanEvent{eventType: addWatchEventType, param: msg}
	for _, v := range r.psubs {
		select {
		case <-v.ctx.Done():
			logger.Infof("[forwarding] tryHandlerWatchEvent index:%d contextId:%s is Done", s.ExecutorChildIndex, v.contextId)
			continue
		case v.out <- e:
			logger.Infof("[forwarding] tryHandlerWatchEvent index:%d contextId:%s sended", s.ExecutorChildIndex, v.contextId)
			continue
		default:
			logger.Infof("[forwarding] tryHandlerWatchEvent index:%d contextId:%s blook", s.ExecutorChildIndex, v.contextId)
			v.out <- e
		}
	}
}

func (s *executorChild) tryPublishChildMsg(result *DecodeResult) {
	info := result.PayloadResult.(*define.GateWayUpPayload)
	if info.SubMsgs != nil && len(info.SubMsgs) > 0 {
		for _, v := range info.SubMsgs {
			if v.Sn == result.NatsMsg.Sn {
				logger.Infof("[forwarding] tryPublishChildMsg 网关sn跟子设备sn不能一样 sn:%s msg:%s", v.Sn, utils.JsonToString(v))
				continue
			}
			dInfo, pInfo, err := s.devices.getDeviceAndProduct(v.Sn)
			if err != nil {
				logger.Infof("[forwarding] tryPublishChildMsg getDeviceAndProduct sn:%s msg:%s err:%s", v.Sn, utils.JsonToString(v), err.Error())
				continue
			}

			if dInfo.PSn != result.NatsMsg.Sn {
				logger.Infof("[forwarding] tryPublishChildMsg 设备sn:%s 设备psn:%s 不是 网关sn:%s 子设备 msg:%s", dInfo.Sn, dInfo.PSn, result.NatsMsg.Sn, utils.JsonToString(v))
			}

			topic := fmt.Sprintf("iot.up.%d.%s.%s.%s.%s", dInfo.Group, pInfo.Pk, dInfo.Sn, result.MsgType, v.Msg.Identify)
			payload, _ := json.Marshal(v.Msg)
			m := &messaging.Message{
				Id:        v.Msg.Id,
				ContextId: v.Msg.ContextId,
				Pk:        pInfo.Pk,
				Sn:        dInfo.Sn,
				Topic:     topic,
				Transform: result.NatsMsg.Transform,
				Protocol:  result.NatsMsg.Protocol,
				Supplier:  "gateway",
				Payload:   payload,
				Created:   time.Now().UnixMilli(),
			}
			err = s.natsClient.Publish(context.Background(), topic, m)
			if err != nil {
				logger.Infof("[forwarding] tryPublishChildMsg Publish sn:%s msg:%s err:%s", v.Sn, utils.JsonToString(v), err.Error())
				continue
			}
		}

	}
}

func (s *executorChild) addWatch(req *addWatchEventReq, outCh chan *ChanEvent) {
	sourceReq := req.watchSourceReq
	msgTypes := make([]string, 0)
	for _, v := range sourceReq.MsgTypes {
		msgTypes = append(msgTypes, getStringWithMsgType(v))
	}

	r := &tokensListResult{}
	s.fetchTokensList(r, sourceReq.Pks)
	s.fetchTokensList(r, sourceReq.Sns)
	s.fetchTokensList(r, msgTypes)
	s.fetchTokensList(r, sourceReq.Codes)

	size := 4
	for i, v := range r.tokensList {
		size = len(v)
		for i2 := size - 1; i2 >= 0; i2-- {
			if v[i2] != "*" {
				// 至少两个星 转>
				if i2 <= size-3 {
					r.tokensList[i] = r.tokensList[i][0 : i2+1]
					r.tokensList[i] = append(r.tokensList[i], ">")
				}
				break
			}
		}
	}

	for _, v := range r.tokensList {
		logger.Infof("[forwarding] addWatch index:%d contextId:%s sub topic:%s", s.ExecutorChildIndex, req.contextId, v)
	}

	subList := make([]*subscription, 0)
	for _, v := range r.tokensList {
		sub := &subscription{
			out:       outCh,
			ctx:       req.ctx,
			tokens:    v,
			contextId: req.contextId,
		}
		err := s.match.Insert(sub)
		if err != nil {
			logger.Errorf("[forwarding] addWatch index:%d contextId:%s match Insert fail :%s", s.ExecutorChildIndex, req.contextId, err.Error())
			continue
		}
		subList = append(subList, sub)
	}

	logger.Infof("[forwarding] addWatch index:%d contextId:%s", s.ExecutorChildIndex, req.contextId)
	s.contextId2SubList[req.contextId] = subList
}

type tokensListResult struct {
	tokensList [][]string
}

func (s *executorChild) fetchTokensList(r *tokensListResult, tokens []string) {
	if r.tokensList == nil {
		r.tokensList = make([][]string, 0)
		r.tokensList = append(r.tokensList, []string{})
	}
	if len(tokens) == 0 {
		for i, v := range r.tokensList {
			v = append(v, "*")
			r.tokensList[i] = v
		}
	} else if len(tokens) == 1 {
		for i, v := range r.tokensList {
			v = append(v, tokens[0])
			r.tokensList[i] = v
		}
	} else {
		tempList := r.tokensList
		r.tokensList = make([][]string, 0)
		for _, v := range tokens {
			for _, v1 := range tempList {
				newTokens := make([]string, len(v1), len(v1))
				copy(newTokens, v1)
				newTokens = append(newTokens, v)
				r.tokensList = append(r.tokensList, newTokens)
			}
		}
	}
}

func (s *executorChild) cancelWatch(contextId string) {
	logger.Infof("[forwarding] cancelWatch contextId:%s start", contextId)
	if v, ok := s.contextId2SubList[contextId]; ok {
		delete(s.contextId2SubList, contextId)
		for _, v2 := range v {
			err := s.match.remove(v2)
			if err != nil {
				logger.Errorf("[forwarding] cancelWatch contextId:%s match remove fail:%s", contextId, err.Error())
			}
		}
	}
}
