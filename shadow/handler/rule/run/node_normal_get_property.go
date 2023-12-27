package run

import "github.com/visonlv/iot-engine/common/define"

type NodeNormalGetProperty struct {
	NodeBase
}

func BuildNodeNormalGetProperty(parantNode NodeIBase, preResult *NodeExecReult, runningContext *RunningContext, runningNode *define.RuleNode) *NodeNormalGetProperty {
	base := BuildNodeBase(parantNode, preResult, runningContext, runningNode)
	return &NodeNormalGetProperty{NodeBase: *base}
}

func (n *NodeNormalGetProperty) Start() error {
	return nil
}
