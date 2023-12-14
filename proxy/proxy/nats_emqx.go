package proxy

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/iot-engine/common/client"
	"github.com/visonlv/iot-engine/common/client/mqtt"
	"github.com/visonlv/iot-engine/common/client/nats"
	"github.com/visonlv/iot-engine/common/proto/messaging"
	"github.com/visonlv/iot-engine/proxy/app"
)

type Nats2Emqx struct {
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

func NewNats2Emqx(clientId, topicPrefix, subFrom, subQueue string) *Nats2Emqx {
	return &Nats2Emqx{
		clientId:    clientId,
		topicPrefix: topicPrefix,
		subFrom:     subFrom,
		subQueue:    subQueue,
		startGroup:  -1,
		endGroup:    -1,
		lock:        new(sync.Mutex),
	}
}

func (p *Nats2Emqx) Start() error {
	pub, err := mqtt.NewPub(app.Cfg.Emqx.Url, app.Cfg.Emqx.Username, app.Cfg.Emqx.Password, p.clientId, "", time.Duration(app.Cfg.Emqx.Timeout)*time.Millisecond)
	if err != nil {
		logger.Error(fmt.Sprintf("[proxy] Nats2Emqx failed to create MQTT sub: %s", err))
		return err
	}

	sub, err := nats.NewSub(app.Cfg.Nats.Url, app.Cfg.Nats.Username, app.Cfg.Nats.Password, p.clientId, p.subQueue)
	if err != nil {
		logger.Errorf("[proxy] Nats2Emqx failed to connect to nats broker: %s", err)
		return err
	}

	p.pub = pub
	p.sub = sub
	return nil
}

func (p *Nats2Emqx) ReloadGroup(startGroup, endGroup int32) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.startGroup == startGroup && p.endGroup == endGroup {
		logger.Infof("[proxy] Nats2Emqx reloadGroup Nats2Emqx group same")
		return
	}

	ctx := context.Background()
	//添加的订阅
	for i := startGroup; i <= endGroup; i++ {
		if i > p.endGroup || i < p.startGroup {
			topic := fmt.Sprintf("%s.%d.>", p.topicPrefix, i)
			err := p.sub.Subscribe(ctx, p.subFrom, topic, p.handle(ctx, p.pub))
			if err != nil {
				logger.Errorf("[proxy] Nats2Emqx reloadGroup Subscribe topic:%s err:%s", topic, err.Error())
			} else {
				logger.Errorf("[proxy] Nats2Emqx reloadGroup Subscribe topic:%s success", topic)
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
					logger.Errorf("[proxy] Nats2Emqx reloadGroup Unsubscribe topic:%s err:%s", topic, err.Error())
				} else {
					logger.Errorf("[proxy] Nats2Emqx reloadGroup Unsubscribe topic:%s success", topic)
				}
			}
		}
	}

	if p.endGroup != -1 && endGroup == -1 {
		logger.Infof("[proxy] Nats2Emqx reloadGroup Unsubscribe 全量取消订阅")
		for i := p.startGroup; i <= p.endGroup; i++ {
			topic := fmt.Sprintf("%s/%d/#", p.topicPrefix, i)
			err := p.sub.Unsubscribe(ctx, p.subFrom, topic)
			if err != nil {
				logger.Errorf("[proxy] Nats2Emqx reloadGroup Unsubscribe topic:%s err:%s", topic, err.Error())
			} else {
				logger.Errorf("[proxy] Nats2Emqx reloadGroup Unsubscribe topic:%s success", topic)
			}
		}
	}
	p.startGroup = startGroup
	p.endGroup = endGroup
}

func (p *Nats2Emqx) handle(ctx context.Context, pub client.Pub) handleFunc {
	return func(msg *messaging.Message) error {
		newTopic := strings.ReplaceAll(msg.Topic, ".", "/")
		if err := pub.Publish(ctx, newTopic, msg); err != nil {
			logger.Warn(fmt.Sprintf("[proxy] Nats2Emqx Failed to forward message: %s", err))
		}
		return nil
	}
}
