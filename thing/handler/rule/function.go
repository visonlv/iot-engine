package rule

import (
	"encoding/json"
	"fmt"

	"github.com/visonlv/iot-engine/common/define"
	"github.com/visonlv/iot-engine/thing/handler/product"
	"github.com/visonlv/iot-engine/thing/model"
)

func Start() error {
	return nil
}

func parseAndGetTrigger(triggerType string, trigger string) (*define.RuleTrigger, error) {
	ruleTrigger := &define.RuleTrigger{}
	err := json.Unmarshal([]byte(trigger), ruleTrigger)
	if err != nil {
		return nil, err
	}

	if triggerType == define.RULE_TRIGGER_TYPE_MSG {
		if ruleTrigger.TriggerMsg == nil {
			return nil, fmt.Errorf("物模型触发参数为空")
		}
		p, err := model.ProductGetByPk(nil, ruleTrigger.TriggerMsg.Pk)
		if err != nil {
			return nil, fmt.Errorf("获取产品:%s 失败:%s", ruleTrigger.TriggerMsg.Pk, err.Error())
		}
		thingInfo, err := product.LoadThingDef(p, true)
		if err != nil {
			return nil, fmt.Errorf("物模型解析失败 产品:%s 失败:%s", ruleTrigger.TriggerMsg.Pk, err.Error())
		}
		ruleTrigger.ThingInfo = thingInfo
	}
	ruleTrigger.Type = triggerType
	ruleTrigger, err = define.RuleTriggerIsValid(ruleTrigger)
	if err != nil {
		return nil, err
	}
	return ruleTrigger, nil
}

func ParseTriggerAndAction(triggerType string, trigger string, action string) (*define.RuleTrigger, *define.RuleNode, error) {
	triggerInfo, err := parseAndGetTrigger(triggerType, trigger)
	if err != nil {
		return nil, nil, err
	}

	node, err := define.ParseRuleRoot(triggerInfo, action)
	if err != nil {
		return nil, nil, err
	}

	return triggerInfo, node, nil
}
