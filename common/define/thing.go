package define

// 物模型类型
type ModelType string

const (
	ModelTypeService  ModelType = "service"
	ModelTypeEvent    ModelType = "event"
	ModelTypeProperty ModelType = "property"
)

// 接入协议
type Protocol string

const (
	ProtocolMqtt3     Protocol = "mqtt3"
	ProtocolMqtt5     Protocol = "mqtt5"
	ProtocolWebsocket Protocol = "websocket"
	ProtocolCoap      Protocol = "coap"
)

// 产品类型
type ProductType string

const (
	ProductTypeDirect  ProductType = "direct"
	ProductTypeGateway ProductType = "gateway"
	ProductTypeChild   ProductType = "child"
)

// 传输类型
type Transform string

const (
	TransformModel Transform = "model"
	TransformRaw   Transform = "raw"
)
