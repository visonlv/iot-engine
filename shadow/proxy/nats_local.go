package proxy

import (
	"context"
	"fmt"
	"sync"

	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/iot-engine/common/client"
	"github.com/visonlv/iot-engine/common/client/nats"
	"github.com/visonlv/iot-engine/common/proto/messaging"
	"github.com/visonlv/iot-engine/shadow/app"
)

type Nats2Local struct {
	clientId    string
	subFrom     string
	topicPrefix string
	subQueue    string
	pub         client.Pub
	sub         client.Sub
	startGroup  int32
	endGroup    int32
	lock        *sync.Mutex
}

func NewNats2Local(clientId, topicPrefix, subFrom, subQueue string) *Nats2Local {
	return &Nats2Local{
		clientId:    clientId,
		topicPrefix: topicPrefix,
		subFrom:     subFrom,
		subQueue:    subQueue,
		startGroup:  -1,
		endGroup:    -1,
		lock:        new(sync.Mutex),
	}
}

func (p *Nats2Local) Start(pub client.Pub) error {
	sub, err := nats.NewSub(app.Cfg.Nats.Url, app.Cfg.Nats.Username, app.Cfg.Nats.Password, p.clientId, p.subQueue)
	if err != nil {
		logger.Errorf("[proxy] Nats2Local failed to connect to nats broker: %s", err)
		return err
	}

	p.pub = pub
	p.sub = sub
	return nil
}

func (p *Nats2Local) ReloadGroup(startGroup, endGroup int32) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.startGroup == startGroup && p.endGroup == endGroup {
		logger.Infof("[proxy] Nats2Local reloadGroup Nats2Local group same")
		return
	}

	ctx := context.Background()
	//添加的订阅
	for i := startGroup; i <= endGroup; i++ {
		if i > p.endGroup || i < p.startGroup {
			topic := fmt.Sprintf("%s.%d.>", p.topicPrefix, i)
			err := p.sub.Subscribe(ctx, p.subFrom, topic, p.handle(ctx, p.pub))
			if err != nil {
				logger.Errorf("[proxy] Nats2Local reloadGroup Subscribe topic:%s err:%s", topic, err.Error())
			} else {
				logger.Errorf("[proxy] Nats2Local reloadGroup Subscribe topic:%s success", topic)
			}
		}
	}
	//取消的订阅
	if p.endGroup != -1 {
		for i := p.startGroup; i <= p.endGroup; i++ {
			if !(i >= startGroup && i <= endGroup) {
				topic := fmt.Sprintf("%s.%d.>", p.topicPrefix, i)
				err := p.sub.Unsubscribe(ctx, p.subFrom, topic)
				if err != nil {
					logger.Errorf("[proxy] Nats2Local reloadGroup Unsubscribe topic:%s err:%s", topic, err.Error())
				} else {
					logger.Errorf("[proxy] Nats2Local reloadGroup Unsubscribe topic:%s success", topic)
				}
			}
		}
	}

	if p.endGroup != -1 && endGroup == -1 {
		logger.Infof("[proxy] Nats2Local reloadGroup Unsubscribe 全量取消订阅")
		for i := p.startGroup; i <= p.endGroup; i++ {
			topic := fmt.Sprintf("%s/%d/#", p.topicPrefix, i)
			err := p.sub.Unsubscribe(ctx, p.subFrom, topic)
			if err != nil {
				logger.Errorf("[proxy] Nats2Local reloadGroup Unsubscribe topic:%s err:%s", topic, err.Error())
			} else {
				logger.Errorf("[proxy] Nats2Local reloadGroup Unsubscribe topic:%s success", topic)
			}
		}
	}
	p.startGroup = startGroup
	p.endGroup = endGroup
}

func (p *Nats2Local) handle(ctx context.Context, pub client.Pub) client.MessageCommonHandler {
	return func(msg *messaging.Message) error {
		if err := pub.Publish(ctx, msg.Topic, msg); err != nil {
			logger.Warn(fmt.Sprintf("[proxy] Nats2Local Failed to forward message: %s", err))
		}
		return nil
	}
}
