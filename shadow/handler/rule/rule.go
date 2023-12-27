package rule

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/go-vkit/utilsx"
	"github.com/visonlv/iot-engine/common/client"
	"github.com/visonlv/iot-engine/common/client/nats"
	"github.com/visonlv/iot-engine/common/define"
	"github.com/visonlv/iot-engine/common/proto/messaging"
	"github.com/visonlv/iot-engine/shadow/app"
	"github.com/visonlv/iot-engine/shadow/handler/forwarding"
	"github.com/visonlv/iot-engine/shadow/handler/rule/run"
	pb "github.com/visonlv/iot-engine/shadow/proto"
	thingpb "github.com/visonlv/iot-engine/thing/proto"
)

const (
	notUseEventType          int = 0
	addOrUpdateRuleEventType int = iota
	deleteRuleEventType
)

type Rule struct {
	id2RuleInfo     map[string]*forwarding.RuleInfo
	id2RuleInfoLock *sync.RWMutex
	sub             client.Sub
	cron            *cron.Cron
}

func newFule() *Rule {
	r := &Rule{
		id2RuleInfo:     make(map[string]*forwarding.RuleInfo),
		id2RuleInfoLock: new(sync.RWMutex),
	}
	return r
}

func (r *Rule) start() error {
	sub, err := nats.NewSub(app.Cfg.Nats.Url, app.Cfg.Nats.Username, app.Cfg.Nats.Password, "rule", "")
	if err != nil {
		logger.Errorf("[proxy] Nats2Local failed to connect to nats broker: %s", err)
		return err
	}
	r.sub = sub

	r.cron = cron.New(cron.WithSeconds())
	r.cron.Start()

	newResp, err := app.Client.RuleService.List(context.Background(), &thingpb.RuleListReq{})
	if err != nil {
		return err
	}

	if newResp.Code != 0 {
		return fmt.Errorf("请求列表错误 code:%d msg:%s", newResp.Code, newResp.Msg)
	}
	for _, v := range newResp.Items {
		err := r.startOneRule(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Rule) tryDelRule(ruleId string, runningId string) error {
	r.id2RuleInfoLock.Lock()
	defer r.id2RuleInfoLock.Unlock()

	ruleInfo, ok := r.id2RuleInfo[ruleId]
	if !ok {
		return nil
	}

	if runningId == "" || ruleInfo.RunningId == runningId {
		delete(r.id2RuleInfo, ruleId)
		logger.Infof("[rule] tryDelRule ruleId:%s runningId:%s", ruleInfo.Id, ruleInfo.RunningId)
		if ruleInfo.CancelFunc != nil {
			ruleInfo.CancelFunc()
			ruleInfo.CancelFunc = nil
		}
		ruleInfo.IsClose = true
	}

	return nil
}

func (r *Rule) addRule(info *forwarding.RuleInfo) error {
	r.id2RuleInfoLock.Lock()
	defer r.id2RuleInfoLock.Unlock()

	r.id2RuleInfo[info.Id] = info
	logger.Infof("[rule] addRule ruleId:%s runningId:%s", info.Id, info.RunningId)

	return nil
}

func (r *Rule) startOneRule(v *thingpb.Rule) error {
	trigger := &define.RuleTrigger{}
	err := json.Unmarshal([]byte(v.Trigger), trigger)
	if err != nil {
		return err
	}

	actionInfo := &define.RuleNode{}
	err = json.Unmarshal([]byte(v.Action), actionInfo)
	if err != nil {
		return err
	}

	ruleInfo := &forwarding.RuleInfo{
		Id:          v.Id,
		Name:        v.Name,
		TriggerType: v.TriggerType,
		Trigger:     v.Trigger,
		TriggerInfo: trigger,
		Action:      v.Action,
		ActionInfo:  actionInfo,
		IsClose:     false,
		RunningId:   utilsx.GenUuid(),
	}

	r.tryDelRule(ruleInfo.Id, ruleInfo.RunningId)

	if v.TriggerType == define.RULE_TRIGGER_TYPE_MSG {
		newReq := &pb.ForwardingWatchReq{
			ContextId: ruleInfo.RunningId,
			Pks:       []string{trigger.TriggerMsg.Pk},
			Sns:       trigger.TriggerMsg.Sns,
			MsgTypes:  []pb.MSG_TYPE(trigger.TriggerMsg.MsgType),
			Codes:     []string{trigger.TriggerMsg.Code},
		}

		go func() {
			defer func() {
				r.tryDelRule(ruleInfo.Id, ruleInfo.RunningId)
			}()
			newCtx, newCancel := context.WithCancel(context.Background())
			ruleInfo.Ctx = newCtx
			ruleInfo.CancelFunc = newCancel

			sBack := func(rsp *forwarding.GetWatchResp) error {
				r.runRule(rsp)
				return nil
			}
			err := forwarding.WatchByRule(newCtx, newReq, ruleInfo, sBack)
			if err != nil {
				r.tryDelRule(ruleInfo.Id, ruleInfo.RunningId)
				logger.Errorf("WatchByRule ruleId:%s fail:%s", ruleInfo.Id, err.Error())
			}
		}()
	} else if v.TriggerType == define.RULE_TRIGGER_TYPE_TIME {
		eid, err := r.cron.AddFunc(ruleInfo.TriggerInfo.TriggerTimer.Cron, func() {
			r.runRule(&forwarding.GetWatchResp{RuleInfo: ruleInfo})
		})

		if err != nil {
			logger.Errorf("WatchByRule ruleId:%s fail:%s", ruleInfo.Id, err.Error())
			return err
		}

		newCtx, newCancel := context.WithCancel(context.Background())
		ruleInfo.Ctx = newCtx
		ruleInfo.CancelFunc = newCancel

		go func() {
			<-ruleInfo.Ctx.Done()
			r.cron.Remove(eid)
		}()
	} else if v.TriggerType == define.RULE_TRIGGER_TYPE_MANUAL {
	} else if v.TriggerType == define.RULE_TRIGGER_TYPE_TOPIC {
		handle := func() client.MessageCommonHandler {
			return func(msg *messaging.Message) error {
				r.runRule(&forwarding.GetWatchResp{RuleInfo: ruleInfo, Msg: msg})
				return nil
			}
		}
		err := r.sub.Subscribe(context.Background(), v.Id, trigger.TriggerTopic.Topic, handle())
		if err != nil {
			r.tryDelRule(ruleInfo.Id, ruleInfo.RunningId)
			return err
		}

		newCtx, newCancel := context.WithCancel(context.Background())
		ruleInfo.Ctx = newCtx
		ruleInfo.CancelFunc = newCancel

		go func() {
			<-ruleInfo.Ctx.Done()
			r.sub.Unsubscribe(context.Background(), v.Id, trigger.TriggerTopic.Topic)
		}()
	} else {
		return fmt.Errorf("not support type:%s", v.TriggerType)
	}

	r.addRule(ruleInfo)
	return nil
}

func (r *Rule) runRule(rsp *forwarding.GetWatchResp) error {
	go func() {
		node := run.StartFirstNode(rsp)
		for {
			select {
			case <-node.GetCtx().Done():
				logger.Infof("任务完成")
				break
			}
		}
	}()
	return nil
}
