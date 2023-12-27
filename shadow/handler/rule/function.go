package rule

import "github.com/visonlv/iot-engine/common/define"

var _p *Rule

func Start() error {
	_p = newFule()
	err := _p.start()
	if err != nil {
		panic(err)
	}
	return nil
}

func RuleConditionModelPropertyKey(r *define.RuleNode) map[string]bool {
	keys := make(map[string]bool)
	for _, v := range r.Conditions {
		tempConddtion := v
		for {
			if tempConddtion.Key.From == define.RULE_CONDITION_KEY_FROM_MODEL {
				keys[tempConddtion.Key.Word[0]] = true
			} else if tempConddtion.Key.From == define.RULE_CONDITION_KEY_FROM_MODEL_PRE {
				keys[tempConddtion.Key.Word[0]] = true
			}
			if v.Next != nil {
				tempConddtion = v.Next
			} else {
				break
			}
		}
	}
	return keys
}
