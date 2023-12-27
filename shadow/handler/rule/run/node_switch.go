package run

import (
	"github.com/visonlv/iot-engine/common/define"
)

type NodeSwitch struct {
	NodeBase
	isSwitch bool
}

func BuildNodeSwitch(parantNode NodeIBase, preResult *NodeExecReult, runningContext *RunningContext, runningNode *define.RuleNode, isSwitch bool) *NodeSwitch {
	base := BuildNodeBase(parantNode, preResult, runningContext, runningNode)
	return &NodeSwitch{NodeBase: *base, isSwitch: isSwitch}
}

func (n *NodeSwitch) Start() error {
	if n.isSwitch {
		n.RunningNode.Nodes = n.RunningNode.Nodes[0:1]
	} else {
		n.RunningNode.Nodes = n.RunningNode.Nodes[1:2]
	}
	n.ShouldFinishChildCount = len(n.RunningNode.Nodes)
	if n.FinishChildCount == n.ShouldFinishChildCount {
		n.SetResult(BuildSuccessResult(n.RunningNode, nil))
		return nil
	}
	n.startNextChild(nil)
	return nil
}

func (n *NodeSwitch) startNextChild(result *NodeExecReult) error {
	curNode := n.RunningNode.Nodes[n.FinishChildCount]
	return CheckConditionAndStart(n, curNode, result)
}

func (n *NodeSwitch) SetResult(r *NodeExecReult) {
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
