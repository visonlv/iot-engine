package run

import (
	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/iot-engine/common/define"
	"github.com/visonlv/iot-engine/shadow/handler/forwarding"
)

func CheckConditionAndStart(node NodeIBase, curNode *define.RuleNode, result *NodeExecReult) error {
	// 判断是否物模型触发
	var thingInfo *define.ThingInfo
	var params map[string]any
	var preShadow *define.Shadow
	if node.GetRunningContext().ruleInfo.TriggerType == define.RULE_TRIGGER_TYPE_MSG {
		d, p, err := forwarding.GetDeviceAndProduct(node.GetRunningContext().msg.Sn)
		if err != nil {
			node.SetResult(BuildFailResult(node.GetRunningNode(), err))
			return nil
		}
		thingInfo = p.ThingInfo
		preShadow = d.Shadow
		params = make(map[string]any)
	}

	ok, err := define.FirstConditionIsMatch(node.GetDataCtx(), thingInfo, params, preShadow, curNode.ConditionPacks)
	if err != nil {
		logger.Errorf("[run] startNextChild FirstConditionIsMatch fail:%s", err)
		node.SetResult(BuildFailResult(node.GetRunningNode(), err))
		return nil
	}
	shouldRun := ok

	isSwitch := false
	if curNode.SwitchConditionPacks != nil {
		ok, err := define.FirstConditionIsMatch(node.GetDataCtx(), thingInfo, params, preShadow, curNode.SwitchConditionPacks)
		if err != nil {
			logger.Errorf("[run] startNextChild FirstConditionIsMatch fail:%s", err)
			node.SetResult(BuildFailResult(node.GetRunningNode(), err))
			return nil
		}
		isSwitch = ok
	}

	if !shouldRun {
		node.SetResult(BuildConditionNotMatch(node.GetRunningNode()))
		return nil
	}

	if curNode.Type == define.RULE_NODE_TYPE_NORMAL {
		GetNorMalNode(nil, nil, node.GetRunningContext(), curNode).Start()
	} else if curNode.Type == define.RULE_NODE_TYPE_SERIAL {
		newNode := BuildNodeSerial(node, result, node.GetRunningContext(), curNode)
		newNode.Start()
	} else if curNode.Type == define.RULE_NODE_TYPE_PARALLEL {
		newNode := BuildNodeParallel(node, result, node.GetRunningContext(), curNode)
		newNode.Start()
	} else if curNode.Type == define.RULE_NODE_TYPE_SWITCH {
		newNode := BuildNodeSwitch(node, result, node.GetRunningContext(), curNode, isSwitch)
		newNode.Start()
	}

	return nil
}

func StartFirstNode(rsp *forwarding.GetWatchResp) NodeIBase {
	ruleInfo := rsp.RuleInfo
	curNode := ruleInfo.ActionInfo
	runningContext := NewRunningContext(ruleInfo, rsp.Msg)
	var base NodeIBase
	if curNode.Type == define.RULE_NODE_TYPE_NORMAL {
		base = GetNorMalNode(nil, nil, runningContext, curNode)
	} else if curNode.Type == define.RULE_NODE_TYPE_SERIAL {
		base = BuildNodeSerial(nil, nil, runningContext, curNode)
	} else if curNode.Type == define.RULE_NODE_TYPE_PARALLEL {
		base = BuildNodeParallel(nil, nil, runningContext, curNode)
	} else if curNode.Type == define.RULE_NODE_TYPE_SWITCH {
		base = BuildNodeSwitch(nil, nil, runningContext, curNode, rsp.SwitchPass)
	}
	go base.Start()
	return base
}

func GetNorMalNode(parantNode NodeIBase, preResult *NodeExecReult, runningContext *RunningContext, curNode *define.RuleNode) NodeIBase {
	switch curNode.NormalType {
	case define.RULE_NODE_TYPE_NORMAL_SET_PROPERTY:
		return BuildNodeNormalSetProperty(parantNode, preResult, runningContext, curNode)
	case define.RULE_NODE_TYPE_NORMAL_GET_PROPERTY:
		return BuildNodeNormalGetProperty(parantNode, preResult, runningContext, curNode)
	case define.RULE_NODE_TYPE_NORMAL_SERVICE:
		return BuildNodeNormalService(parantNode, preResult, runningContext, curNode)
	case define.RULE_NODE_TYPE_NORMAL_NOTIFY:
		return BuildNodeNormalNotify(parantNode, preResult, runningContext, curNode)
	case define.RULE_NODE_TYPE_NORMAL_ALARM:
		return BuildNodeNormalAlarm(parantNode, preResult, runningContext, curNode)
	case define.RULE_NODE_TYPE_NORMAL_CANCEL_ALARM:
		return BuildNodeNormalCancelAlarm(parantNode, preResult, runningContext, curNode)
	case define.RULE_NODE_TYPE_NORMAL_DELAY:
		return BuildNodeNormalDelay(parantNode, preResult, runningContext, curNode)
	}
	return nil
}
