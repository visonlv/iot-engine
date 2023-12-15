package define

import (
	"fmt"
)

const (
	//属性
	MsgTypeProperty = "property"
	//属性回复
	MsgTypePropertyReply = "property_reply"
	//事件
	MsgTypeEvent = "event"
	//服务
	MsgTypeService = "service"
	//服务回复
	MsgTypeServiceReply = "service_reply"
)

type CommonPayload struct {
	Id        string `json:"id"`
	ContextId string `json:"context_id"`
	Time      int64  `json:"time"`
}

func (p *CommonPayload) Validate() error {
	if p.Id == "" {
		return fmt.Errorf("Id is empty")
	}
	if p.Time == 0 {
		return fmt.Errorf("Time is zero")
	}
	return nil
}

// 上行消息
type UpPropertyPayload struct {
	CommonPayload
	Params map[string]any `json:"params"`
}

func (p *UpPropertyPayload) Validate() error {
	err := p.CommonPayload.Validate()
	if err != nil {
		return err
	}
	if p.Params == nil {
		return fmt.Errorf("Params is nil")
	}
	if len(p.Params) <= 0 {
		return fmt.Errorf("Params is empty")
	}
	return nil
}

type UpEventPayload struct {
	CommonPayload
	Type   string         `json:"type"`
	Params map[string]any `json:"params"`
}

func (p *UpEventPayload) Validate() error {
	err := p.CommonPayload.Validate()
	if err != nil {
		return err
	}

	if p.Type != string(EventTypeInfo) && p.Type != string(EventTypeAlert) && p.Type != string(EventTypeFault) {
		return fmt.Errorf("Params is nil")
	}
	if p.Params == nil {
		return fmt.Errorf("event type not support")
	}
	return nil
}

type UpServiceReplyPayload struct {
	CommonPayload
	Code   int32          `json:"code"`
	Msg    string         `json:"msg"`
	Params map[string]any `json:"params"`
}

func (p *UpServiceReplyPayload) Validate() error {
	err := p.CommonPayload.Validate()
	if err != nil {
		return err
	}

	if p.ContextId == "" {
		return fmt.Errorf("ContextId is empty")
	}

	return nil
}

type UpPropertyReplyPayload struct {
	CommonPayload
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
}

func (p *UpPropertyReplyPayload) Validate() error {
	err := p.CommonPayload.Validate()
	if err != nil {
		return err
	}

	if p.ContextId == "" {
		return fmt.Errorf("ContextId is empty")
	}

	return nil
}

type UpBatchPropertyReplyPayload struct {
	CommonPayload
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
}

func (p *UpBatchPropertyReplyPayload) Validate() error {
	err := p.CommonPayload.Validate()
	if err != nil {
		return err
	}

	if p.ContextId == "" {
		return fmt.Errorf("ContextId is empty")
	}

	return nil
}

type UpServicePayload struct {
	CommonPayload
	Params map[string]any `json:"params"`
}

func (p *UpServicePayload) Validate() error {
	err := p.CommonPayload.Validate()
	if err != nil {
		return err
	}

	if p.Params == nil {
		return fmt.Errorf("Params is nil")
	}

	return nil
}

// 属性设置 下行
type DownPropertyPayload struct {
	CommonPayload
	Params map[string]any `json:"params"`
}

func (p *DownPropertyPayload) Validate() error {
	err := p.CommonPayload.Validate()
	if err != nil {
		return err
	}
	if p.Params == nil {
		return fmt.Errorf("Params is nil")
	}
	if len(p.Params) <= 0 {
		return fmt.Errorf("Params is empty")
	}
	return nil
}

type DownServicePayload struct {
	CommonPayload
	Params map[string]any `json:"params"`
}

func (p *DownServicePayload) Validate() error {
	err := p.CommonPayload.Validate()
	if err != nil {
		return err
	}

	if p.Params == nil {
		return fmt.Errorf("Params is nil")
	}

	return nil
}

type DownServiceReplyPayload struct {
	CommonPayload
	Code   int32          `json:"code"`
	Msg    string         `json:"msg"`
	Params map[string]any `json:"params"`
}

func (p *DownServiceReplyPayload) Validate() error {
	err := p.CommonPayload.Validate()
	if err != nil {
		return err
	}

	if p.ContextId == "" {
		return fmt.Errorf("ContextId is empty")
	}

	return nil
}

type GateWayChileMsgPayload struct {
	CommonPayload
	Code     int32  `json:"code"`
	Msg      string `json:"msg"`
	Type     string `json:"type"`
	Identify string `json:"identify"`
	Params   any    `json:"params"`
}

func (p *GateWayChileMsgPayload) Validate() error {
	err := p.CommonPayload.Validate()
	if err != nil {
		return err
	}

	if p.Identify == "" {
		return fmt.Errorf("Identify is empty")
	}

	return nil
}

type GateWayChileMsg struct {
	Sn  string                  `json:"sn"`
	Msg *GateWayChileMsgPayload `json:"msg"`
}

func (p *GateWayChileMsg) Validate() error {
	if p.Sn == "" {
		return fmt.Errorf("Sn is empty")
	}

	if p.Msg == nil {
		return fmt.Errorf("Msg is empty")
	}

	err := p.Msg.Validate()
	if err != nil {
		return err
	}

	return nil
}

type GateWayUpPayload struct {
	CommonPayload
	Code    int32              `json:"code"`
	Msg     string             `json:"msg"`
	SubMsgs []*GateWayChileMsg `json:"sub_msgs"`
}

func (p *GateWayUpPayload) Validate() error {
	err := p.CommonPayload.Validate()
	if err != nil {
		return err
	}

	if p.ContextId == "" {
		return fmt.Errorf("ContextId is empty")
	}

	return nil
}

type GateWayDownChileMsg struct {
	Sn  string         `json:"sn"`
	Msg map[string]any `json:"msg"`
}

type GateWayDownPayload struct {
	CommonPayload
	Code    int32                  `json:"code"`
	Msg     string                 `json:"msg"`
	SubMsgs []*GateWayDownChileMsg `json:"sub_msgs"`
}

func (p *GateWayDownPayload) Validate() error {
	err := p.CommonPayload.Validate()
	if err != nil {
		return err
	}

	if p.ContextId == "" {
		return fmt.Errorf("ContextId is empty")
	}

	return nil
}
