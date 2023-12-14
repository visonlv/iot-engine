package forwarding

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/visonlv/iot-engine/common/define"
	"github.com/visonlv/iot-engine/common/proto/messaging"
	pb "github.com/visonlv/iot-engine/shadow/proto"
	thingpb "github.com/visonlv/iot-engine/thing/proto"

	"github.com/visonlv/go-vkit/logger"
)

func getStringWithMsgType(msgType pb.MSG_TYPE) string {
	switch msgType {
	case pb.MSG_TYPE_PROPERTY:
		return define.MsgTypeProperty
	case pb.MSG_TYPE_PROPERTY_REPLY:
		return define.MsgTypePropertyReply
	case pb.MSG_TYPE_EVENT:
		return define.MsgTypeEvent
	case pb.MSG_TYPE_SERVICE:
		return define.MsgTypeService
	case pb.MSG_TYPE_SERVICE_REPLY:
		return define.MsgTypeServiceReply
	}
	return ""
}

type SendMsgReq struct {
	Sn        string
	ContextId string
	Code      string
	MsgType   string
	Payload   []byte
	Timeout   int32
}

type SendMsgResp struct {
	Code int32
	Msg  string
	// 异步请求不返回消息体
	M *messaging.Message
}

type Products struct {
	pk2Product map[string]*Product
}

func (p *Products) AddOrUpdateProduct(product *Product) error {
	if _, ok := p.pk2Product[product.Pk]; ok {
		p.pk2Product[product.Pk] = product
		logger.Infof("[forwarding] UpdateProduct pk:%s", product.Pk)
		return nil
	}

	p.pk2Product[product.Pk] = product
	logger.Infof("[forwarding] AddProduct pk:%s", product.Pk)
	return nil
}

func (p *Products) DelProduct(product *Product) error {
	delete(p.pk2Product, product.Pk)
	logger.Infof("[forwarding] DelProduct pk:%s", product.Pk)
	return nil
}

func (p *Products) GetProduct(pk string) (*Product, error) {
	if v, ok := p.pk2Product[pk]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("pk:%s not found", pk)
}

type Product struct {
	Pk        string
	Transform define.Transform
	Protocol  define.Protocol
	Type      define.ProductType
	ThingInfo *define.ThingInfo
}

func convertProductPb2Info(v *thingpb.Product) (*Product, error) {
	m := &define.ThingInfo{}
	err := json.Unmarshal([]byte(v.ThingDef), m)
	if err != nil {
		return nil, err
	}

	m.PropertyMap = make(map[string]*define.Property)
	for _, v := range m.Properties {
		m.PropertyMap[v.Code] = v
	}

	m.EventMap = make(map[string]*define.Event)
	for _, v := range m.Events {
		m.EventMap[v.Code] = v
		v.ParamMap = make(map[string]*define.Param)
		for _, v1 := range v.Params {
			v.ParamMap[v1.Code] = v1
		}
	}

	m.UpServiceMap = make(map[string]*define.Service)
	m.DownServiceMap = make(map[string]*define.Service)
	for _, v := range m.Services {
		if v.Dir == define.ServiceDirUp {
			m.UpServiceMap[v.Code] = v
		} else {
			m.DownServiceMap[v.Code] = v
		}
		v.InputMap = make(map[string]*define.Param)
		for _, v1 := range v.Input {
			v.InputMap[v1.Code] = v1
		}

		v.OutputMap = make(map[string]*define.Param)
		for _, v1 := range v.Output {
			v.OutputMap[v1.Code] = v1
		}
	}

	product := &Product{
		Pk:        v.Pk,
		Transform: define.Transform(v.Transform),
		Protocol:  define.Protocol(v.Protocol),
		Type:      define.ProductType(v.Type),
		ThingInfo: m,
	}

	return product, nil
}

type DirtyShadows struct {
	delayEvents map[string]*DelayEvent
}

type DelayEvent struct {
}

type DownMsgContexts struct {
	id2Context map[string]*DownMsgContext
}

func (p *DownMsgContexts) Add(c *DownMsgContext) error {
	if _, ok := p.id2Context[c.contextId]; ok {
		p.id2Context[c.contextId] = c
		panic(fmt.Sprintf("context:%s exist", c.contextId))
	}

	p.id2Context[c.contextId] = c
	logger.Infof("[forwarding] DownMsgContexts Add contextId:%s", c.contextId)
	return nil
}

func (p *DownMsgContexts) Del(contextId string) error {
	delete(p.id2Context, contextId)
	logger.Infof("[forwarding] DownMsgContexts Del contextId:%s", contextId)
	return nil
}

func (p *DownMsgContexts) GetAndDel(contextId string) (*DownMsgContext, error) {
	if v, ok := p.id2Context[contextId]; ok {
		p.Del(contextId)
		return v, nil
	}
	return nil, fmt.Errorf("contextId:%s not found", contextId)
}

type DownMsgContext struct {
	contextId string
	ctx       context.Context
	outCh     chan *ChanEvent
}
