package run

import (
	"time"

	"github.com/visonlv/iot-engine/common/define"
)

type NodeNormalDelay struct {
	NodeBase
}

func BuildNodeNormalDelay(parantNode NodeIBase, preResult *NodeExecReult, runningContext *RunningContext, runningNode *define.RuleNode) *NodeNormalDelay {
	base := BuildNodeBase(parantNode, preResult, runningContext, runningNode)
	return &NodeNormalDelay{NodeBase: *base}
}

func (n *NodeNormalDelay) Start() error {
	go func() {
		time.Sleep(time.Millisecond * time.Duration(n.RunningNode.Action.DelayAction.DelayMs))
		n.SetResult(BuildSuccessResult(n.RunningNode, nil))
	}()
	return nil
}
