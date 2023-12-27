package run

import (
	"context"

	"github.com/visonlv/iot-engine/common/define"
)

type NodeExecReult struct {
	Code         int32 //0成功 -1失败 -2中断
	Msg          string
	Params       map[string]any
	SourceNodeId string
}

func BuildSuccessResult(n *define.RuleNode, params map[string]any) *NodeExecReult {
	return &NodeExecReult{
		Code:         0,
		Msg:          "success",
		Params:       params,
		SourceNodeId: n.Id,
	}
}

func BuildFailResult(n *define.RuleNode, err error) *NodeExecReult {
	return &NodeExecReult{
		Code:         -1,
		Msg:          err.Error(),
		Params:       nil,
		SourceNodeId: n.Id,
	}
}

func BuildStopResult(n *define.RuleNode) *NodeExecReult {
	return &NodeExecReult{
		Code:         -2,
		Msg:          "stop",
		Params:       nil,
		SourceNodeId: n.Id,
	}
}

func BuildConditionNotMatch(n *define.RuleNode) *NodeExecReult {
	return &NodeExecReult{
		Code:         -3,
		Msg:          "not match",
		Params:       nil,
		SourceNodeId: n.Id,
	}
}

type NodeIBase interface {
	SetResult(r *NodeExecReult)
	Start() error
	Stop() error
	GetRunningNode() *define.RuleNode
	GetRunningContext() *RunningContext
	GetDataCtx() map[string]map[string]any
	GetCtx() context.Context
	GetCancelFunc() context.CancelFunc
	SetAllFinishFunc(func(*NodeExecReult) error)
}

type NodeBase struct {
	Ctx                    context.Context
	CancelFunc             context.CancelFunc
	AllFinishFunc          func(*NodeExecReult) error
	ParantNode             NodeIBase
	RunningContext         *RunningContext
	RunningNode            *define.RuleNode
	FinishChildCount       int
	ShouldFinishChildCount int
	HasFinish              bool
	DataCtx                map[string]map[string]any
	FinishResultMap        map[string]*NodeExecReult
}

func BuildNodeBase(parantNode NodeIBase, preResult *NodeExecReult, runningContext *RunningContext, runningNode *define.RuleNode) *NodeBase {
	base := &NodeBase{
		ParantNode:             parantNode,
		RunningContext:         runningContext,
		RunningNode:            runningNode,
		FinishChildCount:       0,
		ShouldFinishChildCount: 0,
		HasFinish:              false,
		DataCtx:                make(map[string]map[string]any),
		FinishResultMap:        make(map[string]*NodeExecReult),
	}

	if parantNode == nil {
		ctx, cancelFunc := context.WithCancel(context.Background())
		base.Ctx = ctx
		base.CancelFunc = cancelFunc
	} else {
		ctx, cancelFunc := context.WithCancel(parantNode.GetCtx())
		base.Ctx = ctx
		base.CancelFunc = cancelFunc
	}

	if preResult != nil {
		base.DataCtx["__pre"] = make(map[string]any)
		base.DataCtx["__pre"]["code"] = preResult.Code
		base.DataCtx["__pre"]["msg"] = preResult.Msg
		base.DataCtx["__pre"]["result"] = preResult.Params
	}
	return base

}

func (n *NodeBase) GetRunningNode() *define.RuleNode {
	return n.RunningNode
}

func (n *NodeBase) GetRunningContext() *RunningContext {
	return n.RunningContext
}

func (n *NodeBase) GetDataCtx() map[string]map[string]any {
	return n.DataCtx
}

func (n *NodeBase) GetCtx() context.Context {
	return n.Ctx
}

func (n *NodeBase) GetCancelFunc() context.CancelFunc {
	return n.CancelFunc
}

func (n *NodeBase) SetAllFinishFunc(a func(*NodeExecReult) error) {
	n.AllFinishFunc = a
}

func (n *NodeBase) SetResult(r *NodeExecReult) {
	if r.SourceNodeId == n.RunningNode.Id {
		if n.HasFinish {
			return
		}
		n.HasFinish = true
		if n.AllFinishFunc != nil {
			n.AllFinishFunc(r)
		}
		n.CancelFunc()
	}
	if n.ParantNode != nil {
		n.ParantNode.SetResult(r)
	}
}

func (n *NodeBase) Start() error {
	return nil
}

func (n *NodeBase) Stop() error {
	n.CancelFunc()
	return nil
}
