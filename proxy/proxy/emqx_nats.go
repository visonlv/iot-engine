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

type ConnectEvent struct {
	Username    string `json:"name,omitempty"`
	Ts          int64  `json:"ts,omitempty"`
	Protocol    string `json:"protocol,omitempty"`
	ConnectedAt int64  `json:"connected_at,omitempty"`
	Clientid    string `json:"clientid,omitempty"`
}

type DisConnectEvent struct {
	Username       string `json:"name,omitempty"`
	Ts             int64  `json:"ts,omitempty"`
	Protocol       string `json:"protocol,omitempty"`
	ConnectedAt    int64  `json:"connected_at,omitempty"`
	DisconnectedAt int64  `json:"disconnected_at,omitempty"`
	Clientid       string `json:"clientid,omitempty"`
}

type Emqx2Nats struct {
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

func NewEmqx2Nats(clientId, topicPrefix, subFrom, subQueue string) *Emqx2Nats {
	return &Emqx2Nats{
		clientId:    clientId,
		topicPrefix: topicPrefix,
		subFrom:     subFrom,
		subQueue:    subQueue,
		startGroup:  -1,
		endGroup:    -1,
		lock:        new(sync.Mutex),
	}
}

func (p *Emqx2Nats) Start() error {
	pub, err := nats.NewPub(app.Cfg.Nats.Url, app.Cfg.Nats.Username, app.Cfg.Nats.Password, p.clientId, p.subQueue)
	if err != nil {
		logger.Errorf("[proxy] Emqx2Nats failed to connect to nats broker: %s", err)
		return err
	}

	sub, err := mqtt.NewSub(app.Cfg.Emqx.Url, app.Cfg.Emqx.Username, app.Cfg.Emqx.Password, p.clientId, "", time.Duration(app.Cfg.Emqx.Timeout)*time.Millisecond)
	if err != nil {
		logger.Error(fmt.Sprintf("[proxy] Emqx2Nats failed to create MQTT sub: %s", err))
		return err
	}

	//处理系统消息
	ctx := context.Background()
	err = sub.Subscribe(ctx, p.subFrom, "$SYS/+/+/clients/+/disconnected", p.handle(ctx, pub))
	err = sub.Subscribe(ctx, p.subFrom, "$SYS/+/+/clients/+/connected", p.handle(ctx, pub))
	if err != nil {
		panic(err)
	}
	p.pub = pub
	p.sub = sub
	return nil
}

func (p *Emqx2Nats) ReloadGroup(startGroup, endGroup int32) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.startGroup == startGroup && p.endGroup == endGroup {
		logger.Infof("[proxy] Emqx2Nats reloadGroup Emqx2Nats group same")
		return
	}

	logger.Infof("[proxy] Emqx2Nats ReloadGroup startGroup:%d endGroup:%d", startGroup, endGroup)
	ctx := context.Background()
	//添加的订阅
	for i := startGroup; i <= endGroup; i++ {
		logger.Infof("[proxy] Emqx2Nats ReloadGroup startGroup:%d endGroup:%d i:%d", startGroup, endGroup, i)

		if i > p.endGroup || i < p.startGroup {
			topic := fmt.Sprintf("%s/%d/#", p.topicPrefix, i)
			logger.Infof("[proxy] Emqx2Nats topic:%s", topic)
			err := p.sub.Subscribe(ctx, p.subFrom, topic, p.handle(ctx, p.pub))
			if err != nil {
				logger.Errorf("[proxy] Emqx2Nats reloadGroup Subscribe topic:%s err:%s", topic, err.Error())
			} else {
				logger.Errorf("[proxy] Emqx2Nats reloadGroup Subscribe topic:%s success", topic)
			}
		}
	}
	//取消的订阅
	if p.endGroup != -1 {
		for i := p.startGroup; i <= p.endGroup; i++ {
			if !(i >= startGroup && i <= endGroup) {
				topic := fmt.Sprintf("%s/%d/#", p.topicPrefix, i)
				err := p.sub.Unsubscribe(ctx, p.subFrom, topic)
				if err != nil {
					logger.Errorf("[proxy] Emqx2Nats reloadGroup Unsubscribe topic:%s err:%s", topic, err.Error())
				} else {
					logger.Errorf("[proxy] Emqx2Nats reloadGroup Unsubscribe topic:%s success", topic)
				}
			}
		}
	}
	if p.endGroup != -1 && endGroup == -1 {
		logger.Infof("[proxy] Emqx2Nats reloadGroup Unsubscribe 全量取消订阅")
		for i := p.startGroup; i <= p.endGroup; i++ {
			topic := fmt.Sprintf("%s/%d/#", p.topicPrefix, i)
			err := p.sub.Unsubscribe(ctx, p.subFrom, topic)
			if err != nil {
				logger.Errorf("[proxy] Emqx2Nats reloadGroup Unsubscribe topic:%s err:%s", topic, err.Error())
			} else {
				logger.Errorf("[proxy] Emqx2Nats reloadGroup Unsubscribe topic:%s success", topic)
			}
		}
	}
	p.startGroup = startGroup
	p.endGroup = endGroup
}

func (p *Emqx2Nats) handle(ctx context.Context, pub client.Pub) handleFunc {
	return func(msg *messaging.Message) error {
		// 处理系统消息 上下线为属性上报
		newTopic := strings.ReplaceAll(msg.Topic, "/", ".")
		if err := pub.Publish(ctx, newTopic, msg); err != nil {
			logger.Warn(fmt.Sprintf("[proxy] Emqx2Nats Failed to forward message: %s", err))
		}
		return nil
	}
}
