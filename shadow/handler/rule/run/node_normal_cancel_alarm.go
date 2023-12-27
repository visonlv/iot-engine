package run

import "github.com/visonlv/iot-engine/common/define"

type NodeNormalCancelAlarm struct {
	NodeBase
}

func BuildNodeNormalCancelAlarm(parantNode NodeIBase, preResult *NodeExecReult, runningContext *RunningContext, runningNode *define.RuleNode) *NodeNormalCancelAlarm {
	base := BuildNodeBase(parantNode, preResult, runningContext, runningNode)
	return &NodeNormalCancelAlarm{NodeBase: *base}
}

func (n *NodeNormalCancelAlarm) Start() error {
	return nil
}
