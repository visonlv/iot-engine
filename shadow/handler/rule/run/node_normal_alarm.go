package run

import (
	"time"

	"github.com/visonlv/iot-engine/common/define"
)

type NodeNormalAlarm struct {
	NodeBase
}

func BuildNodeNormalAlarm(parantNode NodeIBase, preResult *NodeExecReult, runningContext *RunningContext, runningNode *define.RuleNode) *NodeNormalAlarm {
	base := BuildNodeBase(parantNode, preResult, runningContext, runningNode)
	return &NodeNormalAlarm{NodeBase: *base}
}

func (n *NodeNormalAlarm) Start() error {
	go func() {
		time.Sleep(time.Second)
		n.SetResult(BuildSuccessResult(n.RunningNode, nil))
	}()
	return nil
}
