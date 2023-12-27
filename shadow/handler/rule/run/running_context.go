package run

import (
	"time"

	"github.com/visonlv/go-vkit/utilsx"
	"github.com/visonlv/iot-engine/common/proto/messaging"
	"github.com/visonlv/iot-engine/shadow/handler/forwarding"
)

type RunningContext struct {
	runingId   string
	startTime  int64
	status     int32 // 运行 成功 失败
	failReason string
	ruleInfo   *forwarding.RuleInfo
	msg        *messaging.Message
}

func NewRunningContext(ruleInfo *forwarding.RuleInfo, msg *messaging.Message) *RunningContext {
	return &RunningContext{
		runingId:  utilsx.GenUuid(),
		startTime: time.Now().UnixMilli(),
		status:    0,
		ruleInfo:  ruleInfo,
		msg:       msg,
	}
}
