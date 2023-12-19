package define

const (
	TRIGGER_TYPE_MESSAGE      = "message"
	TRIGGER_TYPE_TIME         = "time"
	TRIGGER_TYPE_MANUAL       = "manual"
	ACTION_NODE_TYPE_SWITCH   = "switch"
	ACTION_NODE_TYPE_SERIAL   = "serial"
	ACTION_NODE_TYPE_PARALLEL = "parallel"
)

type RuleShakeLimit struct {
	KeyWords   []string `json:"key_words"` //product,device,msgType,code
	Enabled    bool     `json:"enabled"`
	Interval   int32    `json:"interval"`    //时间限制,单位时间内发生多次告警时,只算一次。单位:秒
	Threshold  int32    `json:"threshold"`   //触发阈值,单位时间内发生n次告警,只算一次。
	AlarmFirst bool     `json:"alarm_first"` //当发生第一次告警时就触发,为false时表示最后一次才触发(告警有延迟,但是可以统计出次数)
}

type RuleTriggerTimer struct {
	//cron表达式
	Cron string `json:"cron"`
}

type RuleTriggerManual struct {
}

type RuleTriggerMessage struct {
	Pks      []string `json:"pks"`
	Sns      []string `json:"sns"`
	MsgTypes []string `json:"msg_types"`
	MsgCodes []string `json:"msg_codes"`
}

type RuleTrigger struct {
	ShakeLimit     *RuleShakeLimit     `json:"shake_limit,omitempty"`
	MessageTrigger *RuleTriggerMessage `json:"trigger_message,omitempty"`
	TimerTrigger   *RuleTriggerTimer   `json:"trigger_timer,omitempty"`
	ManualTrigger  *RuleTriggerManual  `json:"trigger_manual,omitempty"`
}

type RuleActionNotify struct {
	NotifyId string `json:"notify_id"`
}

type RuleActionDelay struct {
	DelayMs int64 `json:"delay_ms"`
}

type RuleActionDevice struct {
	Pks     string         `json:"pks"`
	Sns     []string       `json:"sns"` //不选则所有
	MsgType string         `json:"msgType"`
	Code    string         `json:"code"`
	Params  map[string]any `json:"params"`
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

type RuleCondition struct {
	IsAnd   bool           `json:"is_and"`
	Code    string         `json:"code"` // 空为全局
	KeyWord string         `json:"key_word"`
	Op      string         `json:"op"` //等于 不等于 大于 大于等于 小于 小于等于 在...之间  不在...之间 在...之中 不在...之中 距离当前时间大于...秒 距离当前时间小于...秒
	Value   []any          `json:"value"`
	Next    *RuleCondition `json:"next"`
}

type RuleNode struct {
	Code       string           `json:"code"`
	Type       string           `json:"type"`
	Condition  []*RuleCondition `json:"condition,omitempty"`
	ActionType string           `json:"action_type"`      //并行没有动作类型
	Action     *RuleAction      `json:"action,omitempty"` //并行没有动作
	Nodes      []*RuleNode      `json:"nodes,omitempty"`
}
