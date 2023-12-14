package forwarding

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/visonlv/go-vkit/errorsx"
	"github.com/visonlv/iot-engine/common/utils"
	"github.com/visonlv/iot-engine/shadow/app"
	thingpb "github.com/visonlv/iot-engine/thing/proto"
)

type executor struct {
	executorChildCount int
	executorChildren   []*executorChild
	isReady            int32
}

func newExecutor(executorChildCount int) *executor {
	s := &executor{
		executorChildCount: executorChildCount,
		executorChildren:   make([]*executorChild, 0),
	}
	for i := 0; i < executorChildCount; i++ {
		s.executorChildren = append(s.executorChildren, newExecutorChild(i))
	}
	return s
}

func (s *executor) start() error {
	list, err := s.loadProducts()
	if err != nil {
		return err
	}

	for _, v := range s.executorChildren {
		v.start(list)
	}

	atomic.StoreInt32(&s.isReady, 1)
	return nil
}

func (s *executor) IsReady() bool {
	return atomic.LoadInt32(&s.isReady) == 1
}

func (s *executor) getExecutorChild(sn string) *executorChild {
	slot := utils.GetHashSlotId("inner"+sn, int32(s.executorChildCount))
	return s.executorChildren[slot]
}

func (s *executor) loadProducts() ([]*Product, error) {
	resp, err := app.Client.ProductService.List(context.Background(), &thingpb.ProductListReq{LoadModelDef: true})
	if err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("loadProducts fail code:%d msg:%s", resp.Code, resp.Msg)
	}

	list := make([]*Product, 0)
	for _, v := range resp.Items {
		m, err := convertProductPb2Info(v)
		if err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, nil
}

// interface conversion: interface {} is *proto.Product, not *forwarding.Product

func (s *executor) reloadProduct(isDelete bool, p *thingpb.Product) error {
	m, err := convertProductPb2Info(p)
	if err != nil {
		return err
	}

	if isDelete {
		for _, v := range s.executorChildren {
			v.ayncAddSimpleEvent(deleteProductEventType, m)
		}
		return nil
	}

	for _, v := range s.executorChildren {
		v.ayncAddSimpleEvent(addOrUpdateProductEventType, m)
	}
	return nil
}

func (s *executor) asynSendMsg(msg *SendMsgReq) error {
	child := s.getExecutorChild(msg.Sn)

	event := &ChanInEvent{
		outCh: make(chan *ChanEvent),
		in: &ChanEvent{
			eventType: asynSendMsgEventType,
			param:     msg,
		},
	}
	child.asyncAddEvent(event)
	result := <-event.outCh
	if result.param == nil {
		return nil
	}
	return result.param.(error)
}

func (s *executor) synSendMsg(ctx context.Context, msg *SendMsgReq) (*SendMsgResp, error) {
	child := s.getExecutorChild(msg.Sn)
	event := &ChanInEvent{
		outCh: make(chan *ChanEvent),
		in: &ChanEvent{
			eventType: synSendMsgEventType,
			param:     &synSendMsgEventReq{ctx: ctx, msg: msg},
		},
	}
	child.asyncAddEvent(event)

	select {
	//超时
	case <-ctx.Done():
		m := &SendMsgResp{Code: errorsx.FAIL.Code, Msg: "请求超时"}
		child.asyncAddEvent(&ChanInEvent{
			in: &ChanEvent{
				eventType: cancelContextEventType,
				param:     msg.ContextId,
			}})
		return m, nil
	case result := <-event.outCh:
		resp := result.param.(*synSendMsgEventResp)
		return resp.msg, resp.err
	}
}
