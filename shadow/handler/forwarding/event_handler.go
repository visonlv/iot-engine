package forwarding

import (
	"encoding/json"
	"fmt"

	"github.com/visonlv/iot-engine/common/define"
	"github.com/visonlv/iot-engine/common/proto/messaging"
)

type EventHandler struct {
}

func newEventHandler() *EventHandler {
	return &EventHandler{}
}

type PayloadParamItem struct {
	Defide any
	Value  any
}

type DecodeResult struct {
	MsgType             string
	NatsMsg             *messaging.Message
	Device              *Device
	PayloadResult       any
	PayloadCommonResult define.CommonPayload
}

// 1、物模型校验
// 2、属性存库
// 3、属性订阅唤醒
func (d *EventHandler) handlePropertyMsg(msgType, code string, msg *messaging.Message, product *Product, device *Device) (*DecodeResult, error) {
	info := &define.UpPropertyPayload{}
	err := json.Unmarshal(msg.Payload, info)
	if err != nil {
		return nil, err
	}
	err = info.Validate()
	if err != nil {
		return nil, err
	}

	if info.ContextId != "" {
		msg.ContextId = info.ContextId
	}
	//检查参数属性
	decodeResult := &DecodeResult{
		MsgType:             msgType,
		NatsMsg:             msg,
		PayloadResult:       info,
		PayloadCommonResult: info.CommonPayload,
		Device:              device,
	}

	// 批量属性判断
	if code != define.MsgCodePropertyBatch {
		if len(info.Params) != 1 {
			return nil, fmt.Errorf("model single property:%s must one property set", code)
		}
		if _, ok := info.Params[code]; !ok {
			return nil, fmt.Errorf("model single property:%s property not same to topic", code)
		}
	}

	newParams := make(map[string]any)
	for k, v := range info.Params {
		def, ok := product.ThingInfo.PropertyMap[k]
		if !ok {
			return nil, fmt.Errorf("model not define property key:%s", k)
		}
		baseDef := &def.BaseParamDefine
		value, err := define.ParseVal(baseDef, k, v)
		if err != nil {
			return nil, fmt.Errorf("model parse property k:%s v:%v fail:%v", k, v, err)
		}
		newParams[k] = value
	}

	info.Params = newParams
	return decodeResult, nil
}

// 1、唤醒上下文
// 2、订阅推送
// 3、更新指令对应结果
func (d *EventHandler) handlePropertyReplyMsg(msgType, code string, msg *messaging.Message, product *Product, device *Device) (*DecodeResult, error) {
	info := &define.UpPropertyReplyPayload{}
	err := json.Unmarshal(msg.Payload, info)
	if err != nil {
		return nil, err
	}
	err = info.Validate()
	if err != nil {
		return nil, err
	}

	if info.ContextId != "" {
		msg.ContextId = info.ContextId
	}

	//检查参数属性
	decodeResult := &DecodeResult{
		MsgType:             msgType,
		NatsMsg:             msg,
		PayloadResult:       info,
		PayloadCommonResult: info.CommonPayload,
		Device:              device,
	}

	return decodeResult, nil
}

// 1、警告系统推送
// 2、订阅推送
func (d *EventHandler) handleEventMsg(msgType, code string, msg *messaging.Message, product *Product, device *Device) (*DecodeResult, error) {
	info := &define.UpEventPayload{}
	err := json.Unmarshal(msg.Payload, info)
	if err != nil {
		return nil, err
	}
	err = info.Validate()
	if err != nil {
		return nil, err
	}

	if info.ContextId != "" {
		msg.ContextId = info.ContextId
	}
	//检查参数属性
	decodeResult := &DecodeResult{
		MsgType:             msgType,
		NatsMsg:             msg,
		PayloadResult:       info,
		PayloadCommonResult: info.CommonPayload,
		Device:              device,
	}

	eventDef, ok := product.ThingInfo.EventMap[code]
	if !ok {
		return nil, fmt.Errorf("model not define event:%s", code)
	}

	for k, _ := range info.Params {
		if _, ok := eventDef.ParamMap[k]; !ok {
			return nil, fmt.Errorf("model event param key:%s not exist", k)
		}
	}

	newParams := make(map[string]any)
	for k, v := range eventDef.ParamMap {
		param, ok := info.Params[k]
		if !ok && v.Required {
			return nil, fmt.Errorf("model event:%s param:%s required", code, k)
		}

		if ok {
			baseDef := &v.BaseParamDefine
			value, err := define.ParseVal(baseDef, k, param)
			if err != nil {
				return nil, fmt.Errorf("model parse event:%s param:%s fail:%v", code, k, err)
			}
			newParams[k] = value
		}
	}
	info.Params = newParams
	return decodeResult, nil
}

func (d *EventHandler) handleServiceMsg(msgType, code string, msg *messaging.Message, product *Product, device *Device) (*DecodeResult, error) {
	info := &define.UpServicePayload{}
	err := json.Unmarshal(msg.Payload, info)
	if err != nil {
		return nil, err
	}
	err = info.Validate()
	if err != nil {
		return nil, err
	}

	if info.ContextId != "" {
		msg.ContextId = info.ContextId
	}
	//检查参数属性
	decodeResult := &DecodeResult{
		MsgType:             msgType,
		NatsMsg:             msg,
		PayloadResult:       info,
		PayloadCommonResult: info.CommonPayload,
		Device:              device,
	}

	serviceDef, ok := product.ThingInfo.UpServiceMap[code]
	if !ok {
		return nil, fmt.Errorf("model not define service:%s", code)
	}

	for k, _ := range info.Params {
		if _, ok := serviceDef.InputMap[k]; !ok {
			return nil, fmt.Errorf("model service param key:%s not exist", k)
		}
	}

	newParams := make(map[string]any)
	for k, v := range serviceDef.InputMap {
		param, ok := info.Params[k]
		if !ok && v.Required {
			return nil, fmt.Errorf("model service param key:%s required", k)
		}
		if ok {
			baseDef := &v.BaseParamDefine
			value, err := define.ParseVal(baseDef, k, param)
			if err != nil {
				return nil, fmt.Errorf("model service parse param k:%s param:%v fail:%v", k, param, err.Error())
			}
			newParams[k] = value
		}
	}
	info.Params = newParams
	return decodeResult, nil
}

func (d *EventHandler) handleServiceReplyMsg(msgType, code string, msg *messaging.Message, product *Product, device *Device) (*DecodeResult, error) {
	info := &define.UpServiceReplyPayload{}
	err := json.Unmarshal(msg.Payload, info)
	if err != nil {
		return nil, err
	}
	err = info.Validate()
	if err != nil {
		return nil, err
	}

	if info.ContextId != "" {
		msg.ContextId = info.ContextId
	}
	//检查参数属性
	decodeResult := &DecodeResult{
		MsgType:             msgType,
		NatsMsg:             msg,
		PayloadResult:       info,
		PayloadCommonResult: info.CommonPayload,
		Device:              device,
	}
	serviceDef, ok := product.ThingInfo.DownServiceMap[code]
	if !ok {
		return nil, fmt.Errorf("model not define service:%s", code)
	}

	for k, _ := range info.Params {
		if _, ok := serviceDef.OutputMap[k]; !ok {
			return nil, fmt.Errorf("model service param key:%s not exist", k)
		}
	}

	newParams := make(map[string]any)
	for k, v := range serviceDef.OutputMap {
		param, ok := info.Params[k]
		if !ok && v.Required {
			return nil, fmt.Errorf("model service param key:%s required", k)
		}
		if ok {
			baseDef := &v.BaseParamDefine
			value, err := define.ParseVal(baseDef, k, param)
			if err != nil {
				return nil, fmt.Errorf("model service parse param k:%s v:%v fail:%v", k, param, err.Error())
			}
			newParams[k] = value
		}
	}
	info.Params = newParams

	return decodeResult, nil
}

func (d *EventHandler) decodeUpMsg(msgType, code string, msg *messaging.Message, product *Product, device *Device) (result *DecodeResult, err error) {
	// 统一解析body
	switch msgType {
	case define.MsgTypeProperty:
		return d.handlePropertyMsg(msgType, code, msg, product, device) // 网关ack 上行记录
	case define.MsgTypePropertyReply:
		return d.handlePropertyReplyMsg(msgType, code, msg, product, device) // 不需要ack 上行记录
	case define.MsgTypeEvent:
		return d.handleEventMsg(msgType, code, msg, product, device) // 网关ack 上行记录
	case define.MsgTypeService:
		return d.handleServiceMsg(msgType, code, msg, product, device) //业务ack 下行记录
	case define.MsgTypeServiceReply:
		return d.handleServiceReplyMsg(msgType, code, msg, product, device) //不需要ack 上行记录
	}
	return nil, fmt.Errorf("not support msgType:%s", msgType)
}

func (d *EventHandler) handleDownServiceMsg(msgType, code string, msg *messaging.Message, product *Product, device *Device) (*DecodeResult, error) {
	info := &define.DownServicePayload{}
	err := json.Unmarshal(msg.Payload, info)
	if err != nil {
		return nil, err
	}
	err = info.Validate()
	if err != nil {
		return nil, err
	}

	if info.ContextId != "" {
		msg.ContextId = info.ContextId
	}
	//检查参数属性
	decodeResult := &DecodeResult{
		MsgType:             msgType,
		NatsMsg:             msg,
		PayloadResult:       info,
		PayloadCommonResult: info.CommonPayload,
		Device:              device,
	}

	serviceDef, ok := product.ThingInfo.DownServiceMap[code]
	if !ok {
		return nil, fmt.Errorf("model not define service:%s", code)
	}

	for k, _ := range info.Params {
		if _, ok := serviceDef.InputMap[k]; !ok {
			return nil, fmt.Errorf("model service param key:%s not exist", k)
		}
	}

	newParams := make(map[string]any)
	for k, v := range serviceDef.InputMap {
		param, ok := info.Params[k]
		if !ok && v.Required {
			return nil, fmt.Errorf("model service param key:%s required", k)
		}

		if ok {
			baseDef := &v.BaseParamDefine
			value, err := define.ParseVal(baseDef, k, param)
			if err != nil {
				return nil, fmt.Errorf("model service parse param k:%s param:%v fail:%v", k, param, err.Error())
			}
			newParams[k] = value
		}
	}
	info.Params = newParams
	return decodeResult, nil
}

func (d *EventHandler) handleDownServiceReplyMsg(msgType, code string, msg *messaging.Message, product *Product, device *Device) (*DecodeResult, error) {
	info := &define.DownServiceReplyPayload{}
	err := json.Unmarshal(msg.Payload, info)
	if err != nil {
		return nil, err
	}
	err = info.Validate()
	if err != nil {
		return nil, err
	}

	if info.ContextId != "" {
		msg.ContextId = info.ContextId
	}
	//检查参数属性
	decodeResult := &DecodeResult{
		MsgType:             msgType,
		NatsMsg:             msg,
		PayloadResult:       info,
		PayloadCommonResult: info.CommonPayload,
		Device:              device,
	}
	serviceDef, ok := product.ThingInfo.UpServiceMap[code]
	if !ok {
		return nil, fmt.Errorf("model not define service:%s", code)
	}

	for k, _ := range info.Params {
		if _, ok := serviceDef.OutputMap[k]; !ok {
			return nil, fmt.Errorf("model service param key:%s not exist", k)
		}
	}

	newParams := make(map[string]any)
	for k, v := range serviceDef.OutputMap {
		param, ok := info.Params[k]
		if !ok && v.Required {
			return nil, fmt.Errorf("model service param key:%s required", k)
		}
		if ok {
			baseDef := &v.BaseParamDefine
			value, err := define.ParseVal(baseDef, k, param)
			if err != nil {
				return nil, fmt.Errorf("model service parse param k:%s v:%v fail:%v", k, param, err.Error())
			}
			newParams[k] = value
		}
	}
	info.Params = newParams
	return decodeResult, nil
}

func (d *EventHandler) handleDownPropertyMsg(msgType, code string, msg *messaging.Message, product *Product, device *Device) (*DecodeResult, error) {
	info := &define.DownPropertyPayload{}
	err := json.Unmarshal(msg.Payload, info)
	if err != nil {
		return nil, err
	}
	err = info.Validate()
	if err != nil {
		return nil, err
	}

	if info.ContextId != "" {
		msg.ContextId = info.ContextId
	}
	//检查参数属性
	decodeResult := &DecodeResult{
		MsgType:             msgType,
		NatsMsg:             msg,
		PayloadResult:       info,
		PayloadCommonResult: info.CommonPayload,
		Device:              device,
	}

	// 批量属性判断
	if code != define.MsgCodePropertyBatch {
		if len(info.Params) != 1 {
			return nil, fmt.Errorf("model single property:%s must one property set", code)
		}
		if _, ok := info.Params[code]; !ok {
			return nil, fmt.Errorf("model single property:%s property not same to topic", code)
		}
	}

	newParams := make(map[string]any)
	for k, v := range info.Params {
		def, ok := product.ThingInfo.PropertyMap[k]
		if !ok {
			return nil, fmt.Errorf("model not define property key:%s", k)
		}
		baseDef := &def.BaseParamDefine
		value, err := define.ParseVal(baseDef, k, v)
		if err != nil {
			return nil, fmt.Errorf("model parse property k:%s v:%v fail:%v", k, v, err)
		}
		newParams[k] = value
	}

	info.Params = newParams
	return decodeResult, nil
}

func (d *EventHandler) decodeDownMsg(msgType, code string, msg *messaging.Message, product *Product, device *Device) (result *DecodeResult, err error) {
	// 统一解析body
	switch msgType {
	case define.MsgTypeProperty:
		return d.handleDownPropertyMsg(msgType, code, msg, product, device) // 属性设置 下行记录
	case define.MsgTypeService:
		return d.handleDownServiceMsg(msgType, code, msg, product, device) // 服务调用 下行记录
	case define.MsgTypeServiceReply:
		return d.handleDownServiceReplyMsg(msgType, code, msg, product, device) // 服务调用回复 上行记录
	}
	return nil, fmt.Errorf("not support msgType:%s", msgType)
}

func (d *EventHandler) decodeGatewayUpMsg(msgType, code string, msg *messaging.Message) (*DecodeResult, error) {
	info := &define.GateWayUpPayload{}
	err := json.Unmarshal(msg.Payload, info)
	if err != nil {
		return nil, err
	}
	err = info.Validate()
	if err != nil {
		return nil, err
	}

	if info.SubMsgs != nil {
		for _, v := range info.SubMsgs {
			err = v.Validate()
			if err != nil {
				return nil, err
			}
		}
	}

	//检查参数属性
	decodeResult := &DecodeResult{
		NatsMsg:             msg,
		PayloadResult:       info,
		PayloadCommonResult: info.CommonPayload,
		MsgType:             msgType,
	}

	return decodeResult, nil
}
