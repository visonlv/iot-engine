package mqtt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/go-vkit/utilsx"
	"github.com/visonlv/iot-engine/common/client"
	"github.com/visonlv/iot-engine/common/define"
	"github.com/visonlv/iot-engine/common/proto/messaging"
)

const (
	MQTT_ID  = "emqxmqtt"
	PROTOCAL = "mqtt"
	QOS      = 0
)

var (
	ErrConnect                = errors.New("failed to connect to MQTT broker")
	ErrSubscribeTimeout       = errors.New("failed to subscribe due to timeout reached")
	ErrUnsubscribeTimeout     = errors.New("failed to unsubscribe due to timeout reached")
	ErrUnsubscribeDeleteTopic = errors.New("failed to unsubscribe due to deletion of topic")
	ErrNotSubscribed          = errors.New("not subscribed")
	ErrEmptyTopic             = errors.New("empty topic")
	ErrEmptyID                = errors.New("empty ID")
	errPublishTimeout         = errors.New("failed to publish due to timeout reached")
)

type MqttSystemEvent struct {
	Username       string `json:"name,omitempty"`
	Ts             int64  `json:"ts,omitempty"`
	Protocol       string `json:"protocol,omitempty"`
	ConnectedAt    int64  `json:"connected_at,omitempty"`
	DisconnectedAt int64  `json:"disconnected_at,omitempty"`
	Clientid       string `json:"clientid,omitempty"`
}

type MqttProxy struct {
	address       string
	username      string
	passwod       string
	clientId      string
	queue         string
	timeout       time.Duration
	cc            mqtt.Client
	mu            *sync.RWMutex
	subscriptions map[string]subscription
}

func NewPub(address, username, passwod, clientId, queue string, timeout time.Duration) (client.Pub, error) {
	return NewClient(address, username, passwod, clientId, queue, timeout)
}

func NewSub(address, username, passwod, clientId, queue string, timeout time.Duration) (client.Sub, error) {
	return NewClient(address, username, passwod, clientId, queue, timeout)
}

func NewClient(address, username, passwod, clientId, queue string, timeout time.Duration) (client.Proxy, error) {
	opts := mqtt.NewClientOptions().
		SetUsername(username).
		SetPassword(passwod).
		AddBroker(address).
		SetClientID(clientId)
	cc := mqtt.NewClient(opts)
	token := cc.Connect()
	if token.Error() != nil {
		return nil, token.Error()
	}

	ok := token.WaitTimeout(timeout)
	if !ok {
		return nil, ErrConnect
	}

	if token.Error() != nil {
		return nil, token.Error()
	}

	proxy := &MqttProxy{
		address:       address,
		username:      username,
		passwod:       passwod,
		clientId:      clientId,
		queue:         queue,
		timeout:       timeout,
		cc:            cc,
		mu:            new(sync.RWMutex),
		subscriptions: make(map[string]subscription),
	}

	return proxy, nil
}

func (p *MqttProxy) Publish(ctx context.Context, topic string, msg *messaging.Message) error {
	if topic == "" {
		return ErrEmptyTopic
	}

	logger.Infof("[mqtt] Publish msg topic:%s payload:%s", topic, string(msg.Payload))
	token := p.cc.Publish(topic, QOS, false, msg.Payload)
	if token.Error() != nil {
		logger.Infof("[mqtt] Publish msg topic:%s payload:%s fail:%s", topic, string(msg.Payload), token.Error())
		return token.Error()
	}

	if ok := token.WaitTimeout(p.timeout); !ok {
		logger.Infof("[mqtt] Publish msg topic:%s payload:%s fail:%s", topic, string(msg.Payload), errPublishTimeout.Error())
		return errPublishTimeout
	}

	return nil
}

func (p *MqttProxy) Subscribe(ctx context.Context, id, topic string, handler client.MessageHandler) error {
	if id == "" {
		return ErrEmptyID
	}
	if topic == "" {
		return ErrEmptyTopic
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	s, ok := p.subscriptions[id]
	if ok {
		if _, ok := s.get(topic); ok {
			if err := s.unsubscribe(topic, p.timeout); err != nil {
				logger.Infof("[mqtt] id:%s topic:%s unsubscribe fail:%s", id, topic, err.Error())
				return err
			}
			logger.Infof("[mqtt] id:%s topic:%s unsubscribe success", id, topic)
		}
	} else {
		s = subscription{
			cc:            p.cc,
			topic2Handler: make(map[string]mqtt.MessageHandler),
		}
		p.subscriptions[id] = s
	}

	err := s.subscribe(topic, QOS, p.mqttHandler(handler), p.timeout)
	if err != nil {
		logger.Infof("[mqtt] id:%s topic:%s subscribe fail:%s", id, topic, err.Error())
		return err
	}
	logger.Infof("[mqtt] id:%s topic:%s subscribe success", id, topic)
	return nil
}

func (p *MqttProxy) Unsubscribe(ctx context.Context, id, topic string) error {
	if id == "" {
		return ErrEmptyID
	}
	if topic == "" {
		return ErrEmptyTopic
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	s, ok := p.subscriptions[id]
	if !ok {
		logger.Infof("[mqtt] id:%s topic:%s Unsubscribe fail:%s", id, topic, ErrNotSubscribed.Error())
		return ErrNotSubscribed
	}

	if _, ok := s.get(topic); !ok {
		logger.Infof("[mqtt] id:%s topic:%s Unsubscribe fail:%s", id, topic, ErrNotSubscribed.Error())
		return ErrNotSubscribed
	}

	if err := s.unsubscribe(topic, p.timeout); err != nil {
		logger.Infof("[mqtt] id:%s topic:%s Unsubscribe fail:%s", id, topic, err.Error())
		return err
	}
	p.subscriptions[id] = s

	if len(s.topic2Handler) == 0 {
		delete(p.subscriptions, id)
	}
	return nil
}

func (p *MqttProxy) Close() error {
	p.cc.Disconnect(uint(p.timeout))
	return nil
}

func (p *MqttProxy) mqttHandler(h client.MessageHandler) mqtt.MessageHandler {
	return func(c mqtt.Client, m mqtt.Message) {
		logger.Infof("[mqtt] clientId:%s mqtt read msg topic:%s payload:%s", p.clientId, m.Topic(), string(m.Payload()))
		var msg *messaging.Message
		// 系统消息
		if strings.HasPrefix(m.Topic(), "$SYS") {
			pk, sn, group, event, err := client.DecodeMqttSystemTopic(m.Topic())
			if err != nil {
				logger.Error(fmt.Sprintf("[mqtt] Failed to convert topic:%s err:%s", m.Topic(), err))
				return
			}

			eventInfo := &MqttSystemEvent{}
			json.Unmarshal(m.Payload(), eventInfo)

			msg = &messaging.Message{
				Id:        utilsx.GenUuid(),
				ContextId: "",
				Pk:        pk,
				Sn:        sn,
				Topic:     fmt.Sprintf("iot.up.%s.%s.%s.property.%s", group, pk, sn, define.PropertyOnline),
				Transform: "model",
				Protocol:  eventInfo.Protocol,
				Supplier:  "emqx",
				Created:   time.Now().UnixMilli(),
			}

			payloadInfo := &define.UpPropertyPayload{}
			payloadInfo.Id = msg.Id
			payloadInfo.Params = map[string]any{}
			if event == "connected" {
				payloadInfo.Time = eventInfo.ConnectedAt
				payloadInfo.Params[define.PropertyOnline] = true
			} else {
				payloadInfo.Time = eventInfo.DisconnectedAt
				payloadInfo.Params[define.PropertyOnline] = false
			}
			b, _ := json.Marshal(payloadInfo)
			msg.Payload = b
		} else {
			pk, sn, _, _, _, isUp, err := client.DecodeMqttTopic(m.Topic())
			if err != nil {
				logger.Error(fmt.Sprintf("[mqtt] Failed to convert topic:%s err:%s", m.Topic(), err))
				return
			}
			if !isUp {
				logger.Error(fmt.Sprintf("[mqtt] Failed to convert topic:%s not support down msg", m.Topic()))
				return
			}

			msg = &messaging.Message{
				Id:        utilsx.GenUuid(),
				ContextId: "",
				Pk:        pk,
				Sn:        sn,
				Topic:     strings.ReplaceAll(m.Topic(), "/", "."),
				Transform: "model",
				Protocol:  "mqtt",
				Supplier:  "emqx",
				Payload:   m.Payload(),
				Created:   time.Now().UnixMilli(),
			}
		}

		if err := h.Handle(msg); err != nil {
			logger.Errorf(fmt.Sprintf("[mqtt] Failed to handle mqtt message: %s", err))
		}
	}
}
