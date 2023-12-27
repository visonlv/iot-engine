package define

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cast"
	"github.com/visonlv/go-vkit/logger"
)

const (
	RULE_TRIGGER_TYPE_MSG    = "msg"
	RULE_TRIGGER_TYPE_TIME   = "time"
	RULE_TRIGGER_TYPE_MANUAL = "manual"
	RULE_TRIGGER_TYPE_TOPIC  = "topic"

	RULE_NODE_TYPE_SWITCH   = "switch"
	RULE_NODE_TYPE_NORMAL   = "normal"
	RULE_NODE_TYPE_SERIAL   = "serial"
	RULE_NODE_TYPE_PARALLEL = "parallel"

	RULE_NODE_TYPE_NORMAL_SET_PROPERTY = "normal_set_property"
	RULE_NODE_TYPE_NORMAL_GET_PROPERTY = "normal_get_property"
	RULE_NODE_TYPE_NORMAL_SERVICE      = "normal_service"
	RULE_NODE_TYPE_NORMAL_NOTIFY       = "normal_notify"
	RULE_NODE_TYPE_NORMAL_ALARM        = "normal_alarm"
	RULE_NODE_TYPE_NORMAL_CANCEL_ALARM = "normal_cancel_alarm"
	RULE_NODE_TYPE_NORMAL_DELAY        = "normal_delay"

	RULE_ACTION_TYPE_DEVICE       = "device"
	RULE_ACTION_TYPE_DELAY        = "delay"
	RULE_ACTION_TYPE_NOTIFY       = "notify"
	RULE_ACTION_TYPE_ALARM        = "alarm"
	RULE_ACTION_TYPE_CANCEL_ALARM = "cancel_alarm"
	RULE_CONDITION_OP_E           = "e"   //等于
	RULE_CONDITION_OP_NE          = "ne"  //不等于
	RULE_CONDITION_OP_LT          = "lt"  //小于
	RULE_CONDITION_OP_LTE         = "lte" //小于等于
	RULE_CONDITION_OP_GT          = "gt"  //大于
	RULE_CONDITION_OP_GTE         = "gte" //大于等于
	RULE_CONDITION_OP_IN          = "in"  //在...之中
	RULE_CONDITION_OP_NIN         = "nin" //不在...之中
	RULE_CONDITION_OP_BT          = "bt"  //在...之间
	RULE_CONDITION_OP_NBT         = "nbt" //不在...之间

	RULE_CONDITION_KEY_FROM_MODEL     = "model"
	RULE_CONDITION_KEY_FROM_MODEL_PRE = "model_pre"
	RULE_CONDITION_KEY_FROM_SYS       = "sys"
	RULE_CONDITION_KEY_FROM_PRE       = "pre"

	RULE_CONDITION_KEY_TYPE_INT    = "int"
	RULE_CONDITION_KEY_TYPE_FLOAT  = "float"
	RULE_CONDITION_KEY_TYPE_STRING = "string"
	RULE_CONDITION_KEY_TYPE_BOOL   = "bool"
)

type RuleShakeLimit struct {
	Key        []*RuleConditionKeyWord `json:"key"`
	Enabled    bool                    `json:"enabled"`
	Interval   int32                   `json:"interval"`    //时间限制,单位时间内发生多次告警时,只算一次。单位:秒
	Threshold  int32                   `json:"threshold"`   //触发阈值,单位时间内发生n次告警,只算一次。
	AlarmFirst bool                    `json:"alarm_first"` //当发生第一次告警时就触发,为false时表示最后一次才触发(告警有延迟,但是可以统计出次数)
}

type RuleTriggerTimer struct {
	//cron表达式
	Cron string `json:"cron"`
}

type RuleTriggerManual struct {
}

type RuleTriggerMsg struct {
	Pk      string   `json:"pk"`
	IsAllSn bool     `json:"is_all_sn"`
	Sns     []string `json:"sns"`
	MsgType string   `json:"msg_type"`
	Code    string   `json:"code"`
}

type RuleTriggerTopic struct {
	Topic string `json:"topic"`
}

type RuleTrigger struct {
	Type          string             `json:"-"`
	ThingInfo     *ThingInfo         `json:"-"`
	ShakeLimit    *RuleShakeLimit    `json:"shake_limit,omitempty"`
	TriggerMsg    *RuleTriggerMsg    `json:"trigger_msg,omitempty"`
	TriggerTimer  *RuleTriggerTimer  `json:"trigger_timer,omitempty"`
	TriggerManual *RuleTriggerManual `json:"trigger_manual,omitempty"`
	TriggerTopic  *RuleTriggerTopic  `json:"trigger_topic,omitempty"`
}

type RuleActionNotify struct {
	NotifyId string `json:"notify_id"`
}

type RuleActionDelay struct {
	DelayMs int64 `json:"delay_ms"`
}

type RuleActionDevice struct {
	Pk         string         `json:"pk"`
	Sns        []string       `json:"sns"` //不选则所有
	MsgType    string         `json:"msg_type"`
	Code       string         `json:"code"`
	Params     map[string]any `json:"params"`
	IsSync     bool           `json:"is_sync"`     //是否同步调用
	WaitSecond int            `json:"wait_second"` //同步等待时间
}

type RuleActionAlarm struct {
	AlarmId string `json:"alarm_id"`
}

type RuleAction struct {
	NotifyAction *RuleActionNotify `json:"action_notify,omitempty"`
	DelayAction  *RuleActionDelay  `json:"action_delay,omitempty"`
	DeviceAction *RuleActionDevice `json:"action_device,omitempty"`
	AlarmAction  *RuleActionAlarm  `json:"action_alarm,omitempty"`
	Options      map[string]any    `json:"options"`
}

type RuleConditionKeyWord struct {
	Type string   `json:"type"`
	From string   `json:"from"`
	Word []string `json:"word"`
}

type RuleCondition struct {
	IsAnd  bool                  `json:"is_and"`
	Key    *RuleConditionKeyWord `json:"key"`
	Op     string                `json:"op"`
	Values []any                 `json:"values"`
	Next   *RuleCondition        `json:"next"`
}

type RuleNode struct {
	Id               string           `json:"id"`
	Type             string           `json:"type"`
	NormalType       string           `json:"normal_type"`
	Conditions       []*RuleCondition `json:"conditions,omitempty"`
	SwitchConditions []*RuleCondition `json:"switch_conditions,omitempty"`
	ActionType       string           `json:"action_type"`
	Action           *RuleAction      `json:"action,omitempty"`
	Nodes            []*RuleNode      `json:"nodes,omitempty"`
}

func RuleActionValid(actionType string, ruleAction *RuleAction) (*RuleAction, error) {
	switch actionType {
	case RULE_ACTION_TYPE_DEVICE:
		if ruleAction.DeviceAction != nil {
			return nil, fmt.Errorf("设备动作参数为空")
		}
		if ruleAction.DeviceAction.Pk == "" {
			return nil, fmt.Errorf("必须选择一个产品")
		}
		if ruleAction.DeviceAction.MsgType == "" {
			return nil, fmt.Errorf("必须选择一种输出类型")
		}
		if ruleAction.DeviceAction.Code == "" {
			return nil, fmt.Errorf("必须选择一种指令类型")
		}
		if ruleAction.DeviceAction.Params == nil {
			return nil, fmt.Errorf("必须配置相关指令参数")
		}

		if ruleAction.DeviceAction.IsSync {
			if ruleAction.DeviceAction.WaitSecond <= 0 {
				return nil, fmt.Errorf("同步等待指令需要指定超时时间")
			}
		}
	case RULE_ACTION_TYPE_DELAY:
		if ruleAction.DelayAction.DelayMs <= 0 {
			return nil, fmt.Errorf("延迟动作需要指定超时时间")
		}
	case RULE_ACTION_TYPE_NOTIFY:
		if ruleAction.NotifyAction.NotifyId == "" {
			return nil, fmt.Errorf("通知动作需要指定通知id")
		}
	case RULE_ACTION_TYPE_ALARM:
		if ruleAction.AlarmAction.AlarmId == "" {
			return nil, fmt.Errorf("警告动作需要指定警告id")
		}
	case RULE_ACTION_TYPE_CANCEL_ALARM:
		if ruleAction.AlarmAction.AlarmId == "" {
			return nil, fmt.Errorf("消除警告动作需要指定警告id")
		}
	default:
		return nil, fmt.Errorf("只支持设备输出/延时/通知/警告/消除警告")
	}

	return ruleAction, nil
}

func RuleConditionValid(trigger *RuleTrigger, ruleCondition *RuleCondition) (*RuleCondition, error) {
	if ruleCondition.Key == nil {
		return nil, fmt.Errorf("条件类型跟数据类型参数为空 :%s", jsonToString(ruleCondition))
	}

	values := ruleCondition.Values
	if values == nil || len(values) == 0 {
		return nil, fmt.Errorf("条件参数值为空 :%s", jsonToString(ruleCondition))
	}

	switch ruleCondition.Op {
	case RULE_CONDITION_OP_E:
	case RULE_CONDITION_OP_NE:
	case RULE_CONDITION_OP_LT:
	case RULE_CONDITION_OP_LTE:
	case RULE_CONDITION_OP_GT:
	case RULE_CONDITION_OP_GTE:
	case RULE_CONDITION_OP_IN:
	case RULE_CONDITION_OP_NIN:
	case RULE_CONDITION_OP_BT:
	case RULE_CONDITION_OP_NBT:
	default:
		return nil, fmt.Errorf("不支持该条件操作类型")
	}

	oneValue := values[0]
	newValues := make([]any, 0)
	switch ruleCondition.Key.Type {
	case RULE_CONDITION_KEY_TYPE_INT:
		oneValue, err := cast.ToInt32E(oneValue)
		if err != nil {
			return nil, err
		}
		newValues = append(newValues, oneValue)
		if ruleCondition.Op == RULE_CONDITION_OP_BT || ruleCondition.Op == RULE_CONDITION_OP_NBT {
			if len(values) != 1 {
				return nil, fmt.Errorf("条件必须两个条件参数值 :%s", jsonToString(ruleCondition))
			}
			tempValue, err := cast.ToInt32E(values[1])
			if err != nil {
				return nil, err
			}
			newValues = append(newValues, tempValue)
		}

		if ruleCondition.Op == RULE_CONDITION_OP_IN || ruleCondition.Op == RULE_CONDITION_OP_NIN {
			for index, v := range values {
				if index != 0 {
					tempValue, err := cast.ToInt32E(v)
					if err != nil {
						return nil, err
					}
					newValues = append(newValues, tempValue)
				}
			}
		}
	case RULE_CONDITION_KEY_TYPE_FLOAT:
		oneValue, err := cast.ToFloat32E(oneValue)
		if err != nil {
			return nil, err
		}
		newValues = append(newValues, oneValue)
		if ruleCondition.Op == RULE_CONDITION_OP_BT || ruleCondition.Op == RULE_CONDITION_OP_NBT {
			if len(values) != 1 {
				return nil, fmt.Errorf("条件必须两个条件参数值 :%s", jsonToString(ruleCondition))
			}
			tempValue, err := cast.ToFloat32E(values[1])
			if err != nil {
				return nil, err
			}
			newValues = append(newValues, tempValue)
		}

		if ruleCondition.Op == RULE_CONDITION_OP_IN || ruleCondition.Op == RULE_CONDITION_OP_NIN {
			for index, v := range values {
				if index != 0 {
					tempValue, err := cast.ToFloat32E(v)
					if err != nil {
						return nil, err
					}
					newValues = append(newValues, tempValue)
				}
			}
		}
	case RULE_CONDITION_KEY_TYPE_BOOL:
		if ruleCondition.Op != RULE_CONDITION_OP_E && ruleCondition.Op != RULE_CONDITION_OP_NE {
			return nil, fmt.Errorf("bool数据类型只支持等于跟不等于操作:%s", jsonToString(ruleCondition))
		}
		oneValue, err := cast.ToBoolE(oneValue)
		if err != nil {
			return nil, err
		}
		newValues = append(newValues, oneValue)
	case RULE_CONDITION_KEY_TYPE_STRING:
		if ruleCondition.Op == RULE_CONDITION_OP_E || ruleCondition.Op == RULE_CONDITION_OP_NE {
			oneValue, err := cast.ToStringE(oneValue)
			if err != nil {
				return nil, err
			}
			newValues = append(newValues, oneValue)
		} else if ruleCondition.Op == RULE_CONDITION_OP_IN || ruleCondition.Op == RULE_CONDITION_OP_NIN {
			for _, v := range values {
				tempValue, err := cast.ToStringE(v)
				if err != nil {
					return nil, err
				}
				newValues = append(newValues, tempValue)
			}
		} else {
			return nil, fmt.Errorf("string数据类型只支持等于/不等于/在之中/不在之中操作:%s", jsonToString(ruleCondition))
		}
	}

	switch ruleCondition.Key.From {
	case RULE_CONDITION_KEY_FROM_MODEL:
		//物模型 sn pk dir msg_type code params
	case RULE_CONDITION_KEY_FROM_MODEL_PRE:
		//物模型 sn pk dir msg_type code params
	case RULE_CONDITION_KEY_FROM_SYS:
		//系统预设参数 cur_time(秒级时间戳)
	case RULE_CONDITION_KEY_FROM_PRE:
		//前一个节点结果 code msg result.xxx result.xxx.xxx
	default:
		return nil, fmt.Errorf("不支持该条件参数类型")
	}
	return ruleCondition, nil
}

func RuleNodeIsValid(trigger *RuleTrigger, parentNode *RuleNode, ruleNode *RuleNode) (*RuleNode, error) {
	switch ruleNode.Type {
	case RULE_NODE_TYPE_SWITCH:
		//分支节点没动作
		ruleNode.ActionType = ""
		ruleNode.Action = nil
		newConditions := make([]*RuleCondition, 0)
		if ruleNode.Conditions == nil {
			for _, v := range ruleNode.Conditions {
				newCondition, err := RuleConditionValid(trigger, v)
				if err != nil {
					return nil, err
				}
				newConditions = append(newConditions, newCondition)
			}
		}
		ruleNode.Conditions = newConditions
	case RULE_NODE_TYPE_SERIAL:
		//串行节点没动作
		ruleNode.ActionType = ""
		ruleNode.Action = nil
		newConditions := make([]*RuleCondition, 0)
		if ruleNode.Conditions == nil {
			for _, v := range ruleNode.Conditions {
				newCondition, err := RuleConditionValid(trigger, v)
				if err != nil {
					return nil, err
				}
				newConditions = append(newConditions, newCondition)
			}
		}
		ruleNode.Conditions = newConditions

		newSwitchConditions := make([]*RuleCondition, 0)
		if ruleNode.SwitchConditions == nil {
			for _, v := range ruleNode.SwitchConditions {
				newSerialCondition, err := RuleConditionValid(trigger, v)
				if err != nil {
					return nil, err
				}
				newSwitchConditions = append(newSwitchConditions, newSerialCondition)
			}
		}
		ruleNode.SwitchConditions = newSwitchConditions
	case RULE_NODE_TYPE_PARALLEL:
		//并行节点没动作
		ruleNode.ActionType = ""
		ruleNode.Action = nil
		newConditions := make([]*RuleCondition, 0)
		if ruleNode.Conditions == nil {
			for _, v := range ruleNode.Conditions {
				newCondition, err := RuleConditionValid(trigger, v)
				if err != nil {
					return nil, err
				}
				newConditions = append(newConditions, newCondition)
			}
		}
		ruleNode.Conditions = newConditions
	case RULE_NODE_TYPE_NORMAL:
		if parentNode == nil {
			return nil, fmt.Errorf("普通节点必须有父节点")
		}
		// 检查动作
		action, err := RuleActionValid(ruleNode.ActionType, ruleNode.Action)
		if err != nil {
			return nil, err
		}
		ruleNode.Action = action
		ruleNode.SwitchConditions = nil
		//父节点是分支节点或者并行节点是没有条件的
		if parentNode.Type == RULE_NODE_TYPE_SWITCH || parentNode.Type == RULE_NODE_TYPE_PARALLEL {
			ruleNode.Conditions = nil
		}

		newConditions := make([]*RuleCondition, 0)
		if ruleNode.Conditions == nil {
			for _, v := range ruleNode.Conditions {
				newCondition, err := RuleConditionValid(trigger, v)
				if err != nil {
					return nil, err
				}
				newConditions = append(newConditions, newCondition)
			}
		}
		ruleNode.Conditions = newConditions
	default:
		return nil, fmt.Errorf("只支持分支/串行/并行/普通节点")
	}

	newNodeList := make([]*RuleNode, 0)
	if ruleNode.Nodes != nil {
		for _, v := range ruleNode.Nodes {
			newNode, err := RuleNodeIsValid(trigger, ruleNode, v)
			if err != nil {
				return nil, err
			}
			newNodeList = append(newNodeList, newNode)
		}
	}
	ruleNode.Nodes = newNodeList
	return ruleNode, nil
}

func ParseRuleRoot(trigger *RuleTrigger, rule string) (*RuleNode, error) {
	ruleNode := &RuleNode{}
	err := json.Unmarshal([]byte(rule), ruleNode)
	if err != nil {
		return nil, err
	}
	return RuleNodeIsValid(trigger, nil, ruleNode)
}

func RuleTriggerIsValid(trigger *RuleTrigger) (*RuleTrigger, error) {
	if trigger.ShakeLimit == nil {
		return nil, fmt.Errorf("防抖参数必填")
	}

	if trigger.ShakeLimit.Enabled {
		if trigger.ShakeLimit.Key == nil {
			trigger.ShakeLimit.Key = make([]*RuleConditionKeyWord, 0)
		}

		for _, v := range trigger.ShakeLimit.Key {
			logger.Infof("v:%v", v)
		}
	}

	switch trigger.Type {
	case RULE_TRIGGER_TYPE_MSG:
		if trigger.TriggerMsg == nil {
			return nil, fmt.Errorf("物模型触发参数为空")
		}
		if trigger.TriggerMsg.Pk == "" {
			return nil, fmt.Errorf("物模型触发需要指定产品")
		}
		// 外部检查
		if trigger.TriggerMsg.Pk == "" {
			return nil, fmt.Errorf("物模型触发需要指定产品")
		}

		if !trigger.TriggerMsg.IsAllSn {
			if trigger.TriggerMsg.Sns == nil || len(trigger.TriggerMsg.Pk) <= 0 {
				return nil, fmt.Errorf("物模型触发需要指定设备")
			}
		}

		if !trigger.TriggerMsg.IsAllSn {
			if trigger.TriggerMsg.Sns == nil || len(trigger.TriggerMsg.Pk) <= 0 {
				return nil, fmt.Errorf("物模型触发需要指定设备")
			}
		}

		if trigger.TriggerMsg.MsgType == "" {
			return nil, fmt.Errorf("物模型触发需要指定消息类型")
		}

		if trigger.TriggerMsg.Code == "" {
			return nil, fmt.Errorf("物模型触发需要指定消息代码")
		}
	case RULE_TRIGGER_TYPE_TIME:
		if trigger.TriggerTimer == nil {
			return nil, fmt.Errorf("时间触发参数为空")
		}
		if trigger.TriggerTimer.Cron == "" {
			return nil, fmt.Errorf("时间触发cron时间配置为空")
		}
		//TODO检查cron有效性
	case RULE_TRIGGER_TYPE_MANUAL:
		if trigger.TriggerManual == nil {
			return nil, fmt.Errorf("手动触发参数为空")
		}
	case RULE_TRIGGER_TYPE_TOPIC:
		if trigger.TriggerTopic == nil {
			return nil, fmt.Errorf("主题触发参数为空")
		}
		if trigger.TriggerTopic.Topic == "" {
			return nil, fmt.Errorf("主题触发topic参数为空")
		}
	default:
		return nil, fmt.Errorf("不支持该条件触发类型")
	}

	return trigger, nil
}

func jsonToString(v interface{}) string {
	ret, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(ret)
}

func CheckConditionBool(typeCurValue bool, op, key string, values []any, curValue any) (bool, error) {
	var err error
	if curValue != nil {
		typeCurValue, err = cast.ToBoolE(curValue)
		if err != nil {
			return false, fmt.Errorf("当前属性:%s 并非类型:%s 而是:%T err:%s", key, "bool", curValue, err.Error())
		}
	}
	typeNeedValue, err := cast.ToBoolE(values[0])
	if err != nil {
		return false, fmt.Errorf("要求属性:%s 并非类型:%s 而是:%T err:%s", key, "bool", values[0], err.Error())
	}

	if op == RULE_CONDITION_OP_E {
		return typeCurValue == typeNeedValue, nil
	} else {
		return typeCurValue != typeNeedValue, nil
	}
}

func CheckConditionInt(typeCurValue int32, op, key string, values []any, curValue any) (bool, error) {
	var err error
	if curValue != nil {
		typeCurValue, err = cast.ToInt32E(curValue)
		if err != nil {
			return false, fmt.Errorf("当前属性:%s 并非类型:%s 而是:%T err:%s", key, "int32", curValue, err.Error())
		}
	}
	typeNeedValue, err := cast.ToInt32E(values[0])
	if err != nil {
		return false, fmt.Errorf("要求属性:%s 并非类型:%s 而是:%T err:%s", key, "int32", values[0], err.Error())
	}

	switch op {
	case RULE_CONDITION_OP_E:
		return typeCurValue == typeNeedValue, nil
	case RULE_CONDITION_OP_NE:
		return typeCurValue != typeNeedValue, nil
	case RULE_CONDITION_OP_LT:
		return typeCurValue < typeNeedValue, nil
	case RULE_CONDITION_OP_LTE:
		return typeCurValue <= typeNeedValue, nil
	case RULE_CONDITION_OP_GT:
		return typeCurValue > typeNeedValue, nil
	case RULE_CONDITION_OP_GTE:
		return typeCurValue >= typeNeedValue, nil
	case RULE_CONDITION_OP_IN:
		isContain := false
		for _, v1 := range values {
			typeNeedValue, err := cast.ToInt32E(v1)
			if err != nil {
				return false, fmt.Errorf("要求属性:%s 并非类型:%s 而是:%T err:%s", key, "int32", v1, err.Error())
			}
			if typeCurValue == typeNeedValue {
				isContain = true
				break
			}
		}
		return isContain, nil
	case RULE_CONDITION_OP_NIN:
		isContain := false
		for _, v1 := range values {
			typeNeedValue, err := cast.ToInt32E(v1)
			if err != nil {
				return false, fmt.Errorf("要求属性:%s 并非类型:%s 而是:%T err:%s", key, "int32", v1, err.Error())
			}
			if typeCurValue == typeNeedValue {
				isContain = true
				break
			}
		}
		return !isContain, nil
	case RULE_CONDITION_OP_BT:
		typeNeedValue1, err := cast.ToInt32E(values[1])
		if err != nil {
			return false, fmt.Errorf("要求属性:%s 并非类型:%s 而是:%T err:%s", key, "int32", values[1], err.Error())
		}
		return typeNeedValue1 >= typeCurValue && typeNeedValue <= typeCurValue, nil
	case RULE_CONDITION_OP_NBT:
		typeNeedValue1, err := cast.ToInt32E(values[1])
		if err != nil {
			return false, fmt.Errorf("要求属性:%s 并非类型:%s 而是:%T err:%s", key, "int32", values[1], err.Error())
		}
		return !(typeNeedValue1 >= typeCurValue && typeNeedValue <= typeCurValue), nil
	}
	return false, nil
}

func CheckConditionString(typeCurValue string, op, key string, values []any, curValue any) (bool, error) {
	var err error
	if curValue != nil {
		typeCurValue, err = cast.ToStringE(curValue)
		if err != nil {
			return false, fmt.Errorf("当前属性:%s 并非类型:%s 而是:%T err:%s", key, "string", curValue, err.Error())
		}
	}
	typeNeedValue, err := cast.ToStringE(values[0])
	if err != nil {
		return false, fmt.Errorf("要求属性:%s 并非类型:%s 而是:%T err:%s", key, "string", values[0], err.Error())
	}

	switch op {
	case RULE_CONDITION_OP_E:
		return typeCurValue == typeNeedValue, nil
	case RULE_CONDITION_OP_NE:
		return typeCurValue != typeNeedValue, nil
	case RULE_CONDITION_OP_IN:
		isContain := false
		for _, v1 := range values {
			typeNeedValue, err := cast.ToStringE(v1)
			if err != nil {
				return false, fmt.Errorf("要求属性:%s 并非类型:%s 而是:%T err:%s", key, "string", v1, err.Error())
			}
			if typeCurValue == typeNeedValue {
				isContain = true
				break
			}
		}
		return isContain, nil
	case RULE_CONDITION_OP_NIN:
		isContain := false
		for _, v1 := range values {
			typeNeedValue, err := cast.ToStringE(v1)
			if err != nil {
				return false, fmt.Errorf("要求属性:%s 并非类型:%s 而是:%T err:%s", key, "string", v1, err.Error())
			}
			if typeCurValue == typeNeedValue {
				isContain = true
				break
			}
		}
		return !isContain, nil
	}
	return false, nil
}

func CheckConditionFloat(typeCurValue float32, op, key string, values []any, curValue any) (bool, error) {
	var err error
	if curValue != nil {
		typeCurValue, err = cast.ToFloat32E(curValue)
		if err != nil {
			return false, fmt.Errorf("当前属性:%s 并非类型:%s 而是:%T err:%s", key, "float32", curValue, err.Error())
		}
	}
	typeNeedValue, err := cast.ToFloat32E(values[0])
	if err != nil {
		return false, fmt.Errorf("要求属性:%s 并非类型:%s 而是:%T err:%s", key, "float32", values[0], err.Error())
	}

	switch op {
	case RULE_CONDITION_OP_E:
		return typeCurValue == typeNeedValue, nil
	case RULE_CONDITION_OP_NE:
		return typeCurValue != typeNeedValue, nil
	case RULE_CONDITION_OP_LT:
		return typeCurValue < typeNeedValue, nil
	case RULE_CONDITION_OP_LTE:
		return typeCurValue <= typeNeedValue, nil
	case RULE_CONDITION_OP_GT:
		return typeCurValue > typeNeedValue, nil
	case RULE_CONDITION_OP_GTE:
		return typeCurValue >= typeNeedValue, nil
	case RULE_CONDITION_OP_IN:
		isContain := false
		for _, v1 := range values {
			typeNeedValue, err := cast.ToFloat32E(v1)
			if err != nil {
				return false, fmt.Errorf("要求属性:%s 并非类型:%s 而是:%T err:%s", key, "float32", v1, err.Error())
			}
			if typeCurValue == typeNeedValue {
				isContain = true
				break
			}
		}
		return isContain, nil
	case RULE_CONDITION_OP_NIN:
		isContain := false
		for _, v1 := range values {
			typeNeedValue, err := cast.ToFloat32E(v1)
			if err != nil {
				return false, fmt.Errorf("要求属性:%s 并非类型:%s 而是:%T err:%s", key, "float32", v1, err.Error())
			}
			if typeCurValue == typeNeedValue {
				isContain = true
				break
			}
		}
		return !isContain, nil
	case RULE_CONDITION_OP_BT:
		typeNeedValue1, err := cast.ToFloat32E(values[1])
		if err != nil {
			return false, fmt.Errorf("要求属性:%s 并非类型:%s 而是:%T err:%s", key, "float32", values[1], err.Error())
		}
		return typeNeedValue1 >= typeCurValue && typeNeedValue <= typeCurValue, nil
	case RULE_CONDITION_OP_NBT:
		typeNeedValue1, err := cast.ToFloat32E(values[1])
		if err != nil {
			return false, fmt.Errorf("要求属性:%s 并非类型:%s 而是:%T err:%s", key, "float32", values[1], err.Error())
		}
		return !(typeNeedValue1 >= typeCurValue && typeNeedValue <= typeCurValue), nil
	}
	return false, nil
}

func CheckConditionProperty(propertyDef *BaseParamDefine, op, key string, subKey string, values []any, curValue any) (bool, error) {
	switch propertyDef.Type {
	case DataTypeBool:
		return CheckConditionBool(propertyDef.BoolOptions.Default, op, key, values, curValue)
	case DataTypeInt:
		return CheckConditionInt(propertyDef.IntOptions.Default, op, key, values, curValue)
	case DataTypeString:
		return CheckConditionString("", op, key, values, curValue)
	case DataTypeFloat:
		return CheckConditionFloat(propertyDef.FloatOptions.Default, op, key, values, curValue)
	case DataTypeArray:
		index, err := cast.ToIntE(subKey)
		if err != nil {
			return false, err
		}

		propertySub := propertyDef.ArrayOptions.Array[0]
		if propertyDef.ArrayOptions.Max != 0 && propertyDef.ArrayOptions.Min != 0 {
			if index > int(propertyDef.ArrayOptions.Max) || index < int(propertyDef.ArrayOptions.Min) {
				return false, fmt.Errorf("物模型数组越界:%d", index)
			}
		}
		arr, ok := curValue.([]any)
		if !ok {
			return false, fmt.Errorf("数组类型错误")
		}

		if len(arr) <= index {
			return false, fmt.Errorf("数组越界:%d", index)
		}
		curValue = arr[index]
		return CheckConditionProperty(propertySub, op, subKey, "", values, curValue)
	case DataTypeObject:
		mapkey, err := cast.ToStringE(subKey)
		if err != nil {
			return false, err
		}

		propertySub := propertyDef.ObjectOptions.Object[mapkey]
		curMap, ok := curValue.(map[string]any)
		if !ok {
			return false, fmt.Errorf("字典类型错误")
		}

		curValue, ok = curMap[mapkey]
		if !ok {
			return false, fmt.Errorf("当前值不存在")
		}
		return CheckConditionProperty(propertySub, op, subKey, "", values, curValue)
	default:
		return false, fmt.Errorf("物模型属性数据类型错误:%s", propertyDef.Type)
	}
}

func CheckConditionContext(preContext map[string]any, op, firstKey string, subKey string, values []any) (bool, error) {
	curValue, ok := preContext[firstKey]
	if !ok {
		return false, fmt.Errorf("运行节点参数不存在:%s", firstKey)
	}

	if firstKey == "code" {
		return CheckConditionInt(-1, op, firstKey, values, curValue)
	} else if firstKey == "msg" {
		return CheckConditionString("", op, firstKey, values, curValue)
	} else if firstKey == "result" {
		subValueMap, ok := curValue.(map[string]any)
		if !ok {
			return false, fmt.Errorf("运行节点参数类型错误，不是map[string]any:%s", firstKey)
		}
		subValue, ok := subValueMap[subKey]
		if !ok {
			return false, fmt.Errorf("运行节点参数类型错误，未找到二级key:%s", subKey)
		}
		//尝试浮点型判断
		_, err := cast.ToFloat32E(subValue)
		if err == nil {
			return CheckConditionFloat(0, op, subKey, values, subValue)
		}
		//字符串
		_, err = cast.ToStringE(subValue)
		if err != nil {
			return false, fmt.Errorf("运行节点参数类型错误，二级key的值转字符串异常:%s", subKey)
		}
		return CheckConditionString("", op, subKey, values, subValue)
	} else {
		return false, fmt.Errorf("运行节点参数不存在:%s", firstKey)
	}
}

func CheckOneConditionIsMatch(ctx map[string]map[string]any, thingInfo *ThingInfo, params map[string]any, preShadow *Shadow, ruleCondition *RuleCondition) (bool, error) {
	firstKey := ruleCondition.Key.Word[0]
	secornKey := ""
	if len(ruleCondition.Key.Word) > 1 {
		secornKey = ruleCondition.Key.Word[1]
	}
	switch ruleCondition.Key.From {
	case RULE_CONDITION_KEY_FROM_MODEL:
		//物模型
		propertyDef, ok := thingInfo.PropertyMap[firstKey]
		if !ok {
			return false, fmt.Errorf("物模型没有对应的属性:%s", firstKey)
		}
		curValue, ok := params[firstKey]
		if !ok {
			curValueObj, ok1 := preShadow.Properties[firstKey]
			if ok1 {
				curValue = curValueObj.Current
			}
			ok = ok1
		}
		return CheckConditionProperty(&propertyDef.BaseParamDefine, ruleCondition.Op, firstKey, secornKey, ruleCondition.Values, curValue)
	case RULE_CONDITION_KEY_FROM_MODEL_PRE:
		//物模型
		propertyDef, ok := thingInfo.PropertyMap[firstKey]
		if !ok {
			return false, fmt.Errorf("物模型没有对应的属性:%s", firstKey)
		}
		var curValue any
		curValueObj, ok := preShadow.Properties[firstKey]
		if ok {
			curValue = curValueObj.Current
		}
		return CheckConditionProperty(&propertyDef.BaseParamDefine, ruleCondition.Op, firstKey, secornKey, ruleCondition.Values, curValue)
	case RULE_CONDITION_KEY_FROM_SYS:
		//系统预设参数 cur_time(秒级时间戳)
		if firstKey == "cur_time" {
			return CheckConditionInt(int32(time.Now().Unix()), ruleCondition.Op, firstKey, ruleCondition.Values, nil)
		} else {
			return false, fmt.Errorf("不支持该系统参数:%s", firstKey)
		}
	case RULE_CONDITION_KEY_FROM_PRE:
		//前一个节点结果 code msg result.xxx result.xxx.xxx
		if ctx == nil {
			return false, fmt.Errorf("找不到上下文")
		}
		preContext, ok := ctx["__pre"]
		if !ok {
			return false, fmt.Errorf("未有前置运行节点结果")
		}
		return CheckConditionContext(preContext, ruleCondition.Op, firstKey, secornKey, ruleCondition.Values)
	default:
		return false, fmt.Errorf("不支持该条件参数类型")
	}
}

func FirstConditionIsMatch(ctx map[string]map[string]any, thingInfo *ThingInfo, params map[string]any, preShadow *Shadow, ruleConditions []*RuleCondition) (bool, error) {
	var conditionResult bool
	var oneConditionResult bool
	var oneConditionErr error
	for _, v := range ruleConditions {
		oneConditionResult, oneConditionErr = CheckOneConditionIsMatch(ctx, thingInfo, params, preShadow, v)
		if oneConditionErr != nil {
			return false, oneConditionErr
		}

		if v.Next != nil {
			nextConditionResult, nextConditionErr := CheckOneConditionIsMatch(ctx, thingInfo, params, preShadow, v)
			if nextConditionErr != nil {
				return false, nextConditionErr
			}

			if v.Next.IsAnd {
				oneConditionResult = oneConditionResult && nextConditionResult
			} else {
				oneConditionResult = oneConditionResult || nextConditionResult
			}
		}

		if v.Next.IsAnd {
			conditionResult = conditionResult && oneConditionResult
		} else {
			conditionResult = conditionResult || oneConditionResult
		}
	}
	return conditionResult, nil
}
