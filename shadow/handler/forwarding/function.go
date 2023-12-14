package forwarding

import (
	"context"
	"fmt"
	"time"

	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/iot-engine/common/define"
	"github.com/visonlv/iot-engine/common/proto/messaging"
	pb "github.com/visonlv/iot-engine/shadow/proto"
)

var _p *executor

func Start() error {
	_p = newExecutor(1)
	err := _p.start()
	if err != nil {
		panic(err)
	}
	return nil
}

func GetProperties(req *pb.ForwardingPropertiesReq) ([]*pb.ForwardingProperty, error) {
	var codeMap map[string]string
	if req.Codes != nil && len(req.Codes) > 0 {
		codeMap = map[string]string{}
		for _, v := range req.Codes {
			codeMap[v] = v
		}
	}

	// 多个节点请求并合并
	enterChild2Sns := make(map[int][]string)
	if len(req.Pks) > 0 || len(req.Sns) == 0 {
		for index, _ := range _p.executorChildren {
			enterChild2Sns[index] = []string{}
		}
	}
	if len(req.Sns) > 0 {
		for _, v := range req.Sns {
			child := _p.getExecutorChild(v)
			sns, ok := enterChild2Sns[child.ExecutorChildIndex]
			if !ok {
				sns = make([]string, 0)
			}
			sns = append(sns, v)
			enterChild2Sns[child.ExecutorChildIndex] = sns
		}
	}

	resp := make([]*pb.ForwardingProperty, 0)
	for index, sns := range enterChild2Sns {
		child := _p.executorChildren[index]
		event := &ChanInEvent{
			outCh: make(chan *ChanEvent),
			in: &ChanEvent{
				eventType: getDeviceListEventType,
				param:     sns,
			},
		}
		child.asyncAddEvent(event)
		result := <-event.outCh
		if result != nil && result.param != nil {
			devices := result.param.([]*Device)
			for _, v := range devices {
				propertyMap := make(map[string]*pb.ForwardingPropertyItem)
				for propertyk, propertyv := range v.Shadow.Properties {
					if codeMap != nil {
						if _, ok := codeMap[propertyk]; !ok {
							continue
						}
					}

					newStr, err := define.PropertyCurrentValurAsString(propertyv.Current.Value)
					if err != nil {
						return nil, fmt.Errorf("影子属性格式化错误:%s", err.Error())
					}

					propertyMap[propertyk] = &pb.ForwardingPropertyItem{
						Value:      newStr,
						UpdateTime: propertyv.Current.UpdatedTime,
					}
				}

				resp = append(resp, &pb.ForwardingProperty{
					Pk:          v.Pk,
					Sn:          v.Sn,
					PropertyMap: propertyMap,
				})
			}
		}
	}
	return resp, nil
}

func Watch(ctx context.Context, req *pb.ForwardingWatchReq, s *pb.ForwardingService_WatchServer) error {
	//广播
	children := make([]*executorChild, 0)
	if len(req.Pks) > 0 || len(req.Sns) == 0 {
		children = _p.executorChildren
	}

	//节点
	if len(req.Sns) > 0 {
		isEnterMap := make(map[int]*executorChild)
		for _, v := range req.Sns {
			child := _p.getExecutorChild(v)
			if _, ok := isEnterMap[child.ExecutorChildIndex]; !ok {
				isEnterMap[child.ExecutorChildIndex] = child
				children = append(children, child)
			}
		}
	}

	childCtx, cancelFunc := context.WithCancel(ctx)
	contextId := req.ContextId
	event := &ChanInEvent{
		outCh: make(chan *ChanEvent, 128),
		in: &ChanEvent{
			eventType: addWatchEventType,
			param: &addWatchEventReq{
				contextId:      contextId,
				ctx:            childCtx,
				watchSourceReq: req,
			},
		},
	}

	for _, v := range children {
		v.asyncAddEvent(event)
	}

	defer func() {
		logger.Infof("[forwarding] Watch contextId:%s finish", contextId)
		cancelFunc()
		//移除watch事件
		event := &ChanInEvent{
			outCh: make(chan *ChanEvent),
			in: &ChanEvent{
				eventType: cancelWatchEventType,
				param:     contextId,
			},
		}
		for _, v := range children {
			v.asyncAddEvent(event)
		}
	}()

	logger.Infof("[forwarding] Watch contextId:%s", contextId)

	for {
		select {
		case <-ctx.Done():
			logger.Infof("[forwarding] Watch contextId:%s Done", contextId)
			return nil
		case e := <-event.outCh:
			msg := e.param.(*messaging.Message)
			err := s.Send(&pb.ForwardingWatchResp{M: msg})
			if err != nil {
				logger.Infof("[forwarding] Watch contextId:%s Send err:%s", contextId, err)
				return err
			}
		}
	}
}

func CommonSendMsg(ctx context.Context, msg *SendMsgReq) (*SendMsgResp, error) {
	if _p == nil || !_p.IsReady() {
		logger.Errorf("executor not ready")
		return nil, fmt.Errorf("executor not ready")
	}
	// 异步消息
	if msg.ContextId == "" {
		err := _p.asynSendMsg(msg)
		if err != nil {
			return nil, err
		}
		return &SendMsgResp{}, nil
	}

	// 同步消息
	timeount := 10 * time.Second
	if msg.Timeout > 0 {
		timeount = time.Duration(msg.Timeout) * time.Millisecond
	}

	ctx1, cancel := context.WithTimeout(ctx, timeount)
	defer func() {
		cancel()
	}()
	m, err := _p.synSendMsg(ctx1, msg)
	return m, err

}

func handlerNatsMsg(topic string, msg *messaging.Message) error {
	if _p == nil || !_p.IsReady() {
		logger.Errorf("executor not ready")
		return fmt.Errorf("executor not ready")
	}
	_p.getExecutorChild(msg.Sn).ayncAddSimpleEvent(natsEventType, msg)
	return nil
}

func GetProductByPks(pks []string) (map[string]*Product, error) {
	event := &ChanInEvent{
		outCh: make(chan *ChanEvent),
		in: &ChanEvent{
			eventType: getProductMapEventType,
			param:     pks,
		},
	}

	_p.executorChildren[0].asyncAddEvent(event)
	result := <-event.outCh
	resp := result.param.(*getProductMapResp)
	return resp.pk2Product, resp.err
}
