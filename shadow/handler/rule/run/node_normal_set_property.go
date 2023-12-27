package run

import "github.com/visonlv/iot-engine/common/define"

type NodeNormalSetProperty struct {
	NodeBase
}

func BuildNodeNormalSetProperty(parantNode NodeIBase, preResult *NodeExecReult, runningContext *RunningContext, runningNode *define.RuleNode) *NodeNormalSetProperty {
	base := BuildNodeBase(parantNode, preResult, runningContext, runningNode)
	return &NodeNormalSetProperty{NodeBase: *base}
}

func (n *NodeNormalSetProperty) Start() error {
	return nil
}
