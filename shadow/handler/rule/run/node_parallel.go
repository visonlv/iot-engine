package run

import (
	"github.com/visonlv/iot-engine/common/define"
)

type NodeParallel struct {
	NodeBase
}

func BuildNodeParallel(parantNode NodeIBase, preResult *NodeExecReult, runningContext *RunningContext, runningNode *define.RuleNode) *NodeParallel {
	base := BuildNodeBase(parantNode, preResult, runningContext, runningNode)
	return &NodeParallel{NodeBase: *base}
}

func (n *NodeParallel) Start() error {
	n.ShouldFinishChildCount = len(n.RunningNode.Nodes)
	if n.FinishChildCount == n.ShouldFinishChildCount {
		n.SetResult(BuildSuccessResult(n.RunningNode, nil))
		return nil
	}

	for _, curNode := range n.RunningNode.Nodes {
		CheckConditionAndStart(n, curNode, nil)
	}
	return nil
}

func (n *NodeParallel) SetResult(r *NodeExecReult) {
	n.NodeBase.SetResult(r)
	if r.SourceNodeId != n.RunningNode.Id {
		if _, ok := n.FinishResultMap[r.SourceNodeId]; !ok {
			n.FinishResultMap[r.SourceNodeId] = r
			n.FinishChildCount++
			if n.FinishChildCount == n.ShouldFinishChildCount {
				n.SetResult(BuildSuccessResult(n.RunningNode, nil))
			}
		}
	}
}
