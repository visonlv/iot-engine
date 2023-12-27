package run

import (
	"github.com/visonlv/iot-engine/common/define"
)

type NodeSerial struct {
	NodeBase
}

func BuildNodeSerial(parantNode NodeIBase, preResult *NodeExecReult, runningContext *RunningContext, runningNode *define.RuleNode) *NodeSerial {
	base := BuildNodeBase(parantNode, preResult, runningContext, runningNode)
	return &NodeSerial{NodeBase: *base}
}

func (n *NodeSerial) Start() error {
	n.ShouldFinishChildCount = len(n.RunningNode.Nodes)
	if n.FinishChildCount == n.ShouldFinishChildCount {
		n.SetResult(BuildSuccessResult(n.RunningNode, nil))
		return nil
	}
	n.startNextChild(nil)
	return nil
}

func (n *NodeSerial) startNextChild(result *NodeExecReult) error {
	curNode := n.RunningNode.Nodes[n.FinishChildCount]
	return CheckConditionAndStart(n, curNode, result)
}

func (n *NodeSerial) SetResult(r *NodeExecReult) {
	n.NodeBase.SetResult(r)
	if r.SourceNodeId != n.RunningNode.Id {
		if _, ok := n.FinishResultMap[r.SourceNodeId]; !ok {
			n.FinishResultMap[r.SourceNodeId] = r
			n.FinishChildCount++
			if n.FinishChildCount != n.ShouldFinishChildCount {
				n.startNextChild(r)
			} else {
				n.SetResult(BuildSuccessResult(n.RunningNode, nil))
			}
		}
	}
}
