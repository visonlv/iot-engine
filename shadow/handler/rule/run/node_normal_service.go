package run

import "github.com/visonlv/iot-engine/common/define"

type NodeNormalService struct {
	NodeBase
}

func BuildNodeNormalService(parantNode NodeIBase, preResult *NodeExecReult, runningContext *RunningContext, runningNode *define.RuleNode) *NodeNormalService {
	base := BuildNodeBase(parantNode, preResult, runningContext, runningNode)
	return &NodeNormalService{NodeBase: *base}
}

func (n *NodeNormalService) Start() error {
	return nil
}
