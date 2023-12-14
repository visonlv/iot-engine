package main

import (
	"time"
)

// 接入协议
type Protocol int

const (
	MQTT3 Protocol = iota
	MQTT5
	WEBSOCKET
	COAP
	TCP
	UDP
)

// 产品类型
type ProductType uint8

const (
	DIRECT ProductType = iota
	GATEWAY
	CHILD
)

// 属性类型
type PropertyType uint8

const (
	INT PropertyType = iota
	FLOAT
	BOOL
	STRING
)

// 物模型
type Model struct {
	Properties map[string]PropertyDef
	Services   map[string]ServiceDef
	Events     map[string]EventDef
}

// 属性定义
type PropertyDef struct {
	Name        string
	Description string
	Unit        string
	Type        PropertyType
	Writable    bool
	Min         int16
	Max         int16
	Options     []interface{}
}

// 服务定义
type ServiceDef struct {
	Name        string
	Description string
	Parameters  []ServiceParameterDef
}

// 服务参数定义
type ServiceParameterDef struct {
	Name        string
	Description string
	Type        PropertyType
	Min         int16
	Max         int16
	Options     []interface{}
}

// 事件定义
type EventDef struct {
	Name        string
	Description string
	Parameters  []EventParameterDef
}

// 事件参数定义
type EventParameterDef struct {
	Name        string
	Description string
	Type        PropertyType
}

// 产品
type Product struct {
	Id          string
	Name        string
	Description string
	Protocol    Protocol
	Type        ProductType
	Model       Model
}

// 设备
type Device struct {
	Id      string
	Name    string
	Sn      string
	Key     string
	Product Product
	Shadow  Shadow
}

// 影子
type Shadow struct {
	ConnectionState ConnectionState
	Properties      map[string]PropertyState
}

// 影子行为接口，由Device实现
type ShadowBehavior interface {
	GetPropertyState(propertyName string) *PropertyState
	GetAllPropertyState() map[string]PropertyState
	GetConnectionState() *ConnectionState
	UpdateDesiredState(propertyName string, desiredState *DesiredState)
	Sync() error
}

// 连接状态
type ConnectionState struct {
	State            string
	StateUpdatedTime time.Time
	LastOnlineTime   time.Time
	LastOfflineTime  time.Time
}

// 属性状态
type PropertyState struct {
	Current CurrentState
	Desired DesiredState
}

// 属性当前状态
type CurrentState struct {
	Value       interface{}
	UpdatedTime time.Time
}

// 属性期望状态
type DesiredState struct {
	Value        interface{}
	ReceivedTime time.Time
	Expiration   uint32
}
