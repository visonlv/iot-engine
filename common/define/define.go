package define

const (
	PropertyOnline = "online"
)

const (
	MsgCodeGatewayProxy  = "proxy"
	MsgCodePropertyBatch = "batch"
)

var (
	ModelExceptKeyWord = map[string]bool{
		MsgCodePropertyBatch: true,
		MsgCodeGatewayProxy:  true,
	}
)

func IsSysModelCode(code string) bool {
	_, ok := ModelExceptKeyWord[code]
	return ok
}
