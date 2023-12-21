package define

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cast"
	"github.com/visonlv/go-vkit/logger"
)

const (
	RULE_TRIGGER_TYPE_MODEL  = "model"
	RULE_TRIGGER_TYPE_TIME   = "time"
	RULE_TRIGGER_TYPE_MANUAL = "manual"
	RULE_TRIGGER_TYPE_TOPIC  = "topic"

	RULE_NODE_TYPE_SWITCH   = "switch"
	RULE_NODE_TYPE_NORMAL   = "normal"
	RULE_NODE_TYPE_SERIAL   = "serial"
	RULE_NODE_TYPE_PARALLEL = "parallel"

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
	RULE_CONDITION_KEY_FROM_PARENT    = "parent"

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

type RuleTriggerModel struct {
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
	TriggerModel  *RuleTriggerModel  `json:"trigger_model,omitempty"`
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
	Type string `json:"type"`
	From string `json:"from"`
	Word string `json:"word"`
}

type RuleCondition struct {
	IsAnd  bool                  `json:"is_and"`
	Key    *RuleConditionKeyWord `json:"key"`
	Op     string                `json:"op"`
	Values []any                 `json:"values"`
	Next   *RuleCondition        `json:"next"`
}

type RuleNode struct {
	Id               string                    `json:"id"`
	Type             string                    `json:"type"`
	Conditions       []*RuleCondition          `json:"conditions,omitempty"`
	SerialConditions []*RuleCondition          `json:"serial_conditions,omitempty"`
	ActionType       string                    `json:"action_type"`
	Action           *RuleAction               `json:"action,omitempty"`
	Nodes            []*RuleNode               `json:"nodes,omitempty"`
	ExecContext      map[string]map[string]any `json:"-"`
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
		} else if ruleCondition.Op == RULE_CONDITION_OP_E || ruleCondition.Op == RULE_CONDITION_OP_NE {
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
	case RULE_CONDITION_KEY_FROM_PARENT:
		//父节点结果 code msg result.xxx result.xxx.xxx
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

		newSerialConditions := make([]*RuleCondition, 0)
		if ruleNode.SerialConditions == nil {
			for _, v := range ruleNode.SerialConditions {
				newSerialCondition, err := RuleConditionValid(trigger, v)
				if err != nil {
					return nil, err
				}
				newSerialConditions = append(newSerialConditions, newSerialCondition)
			}
		}
		ruleNode.SerialConditions = newSerialConditions
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
		ruleNode.SerialConditions = nil
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
			newNode.ExecContext = ruleNode.ExecContext
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

	ruleNode.ExecContext = make(map[string]map[string]any)
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
	case RULE_TRIGGER_TYPE_MODEL:
		if trigger.TriggerModel == nil {
			return nil, fmt.Errorf("物模型触发参数为空")
		}
		if trigger.TriggerModel.Pk == "" {
			return nil, fmt.Errorf("物模型触发需要指定产品")
		}
		// 外部检查
		if trigger.TriggerModel.Pk == "" {
			return nil, fmt.Errorf("物模型触发需要指定产品")
		}

		if !trigger.TriggerModel.IsAllSn {
			if trigger.TriggerModel.Sns == nil || len(trigger.TriggerModel.Pk) <= 0 {
				return nil, fmt.Errorf("物模型触发需要指定设备")
			}
		}

		if !trigger.TriggerModel.IsAllSn {
			if trigger.TriggerModel.Sns == nil || len(trigger.TriggerModel.Pk) <= 0 {
				return nil, fmt.Errorf("物模型触发需要指定设备")
			}
		}

		if trigger.TriggerModel.MsgType == "" {
			return nil, fmt.Errorf("物模型触发需要指定消息类型")
		}

		if trigger.TriggerModel.Code == "" {
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
