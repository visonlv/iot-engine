package nats

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/go-vkit/natsx"
	"github.com/visonlv/iot-engine/common/client"
	"github.com/visonlv/iot-engine/common/proto/messaging"
	"google.golang.org/protobuf/proto"
)

var (
	ErrNotSubscribed          = errors.New("not subscribed")
	ErrEmptyTopic             = errors.New("empty topic")
	ErrEmptyID                = errors.New("empty id")
	ErrUnsubscribeDeleteTopic = errors.New("failed to unsubscribe due to deletion of topic")
)

type NatsClient struct {
	url           string
	username      string
	passwod       string
	clientId      string
	queue         string
	cc            *natsx.NatsClient
	mu            *sync.RWMutex
	subscriptions map[string]subscription
}

func NewPub(url, username, passwod, clientId, queue string) (client.Pub, error) {
	return NewClient(url, username, passwod, clientId, queue)
}

func NewSub(url, username, passwod, clientId, queue string) (client.Sub, error) {
	return NewClient(url, username, passwod, clientId, queue)
}

func NewClient(url, username, passwod, clientId, queue string) (*NatsClient, error) {
	cli := &NatsClient{}
	sub := func() {
		if cli.subscriptions == nil {
			logger.Infof("[nats] id:%s cli.subscriptions is nill", clientId)
			return
		}
		for clientId, subscription := range cli.subscriptions {
			for topic, topicInfo := range subscription.topic2Handler {
				logger.Infof("[nats] reconnect sub clientId:%s topic:%s", clientId, topic)
				var sub *nats.Subscription
				var err error
				if queue != "" {
					sub, err = cli.cc.QueueSubscribe(topic, queue, topicInfo.h)
				} else {
					sub, err = cli.cc.Subscribe(topic, topicInfo.h)
				}
				if err != nil {
					logger.Infof("[nats] reconnect sub clientId:%s topic:%s fail err:%s", clientId, topic, err)
				} else {
					topicInfo.sub = sub
				}
			}
		}
	}
	logger.Infof("nast info Url:%s, Username:%s, Password:%s", url, username, passwod)
	newOptions := make([]nats.Option, 0)
	newOptions = append(newOptions, nats.ReconnectHandler(func(c *nats.Conn) {
		logger.Infof("[nats] ReconnectHandler")
	}))

	newOptions = append(newOptions, nats.DisconnectErrHandler(func(c *nats.Conn, e error) {
		logger.Infof("[nats] DisconnectErrHandler")
	}))

	newOptions = append(newOptions, nats.ClosedHandler(func(c *nats.Conn) {
		logger.Infof("[nats] ClosedHandler")
	}))

	cc := natsx.NewNatsClient(url, username, passwod, newOptions...)
	cc.SetConnectedHandler(func(c *nats.Conn) {
		logger.Infof("[nats] nats ConnectedHandler")
		sub()
	})
	sub()

	cli.url = url
	cli.username = username
	cli.passwod = passwod
	cli.clientId = clientId
	cli.queue = queue
	cli.cc = cc
	cli.mu = new(sync.RWMutex)
	cli.subscriptions = make(map[string]subscription)

	return cli, nil
}

func (p *NatsClient) Publish(ctx context.Context, topic string, msg *messaging.Message) error {
	if topic == "" {
		logger.Infof("[nats] Publish msg topic:%s payload:%s fail:%s", topic, string(msg.Payload), ErrEmptyTopic.Error())
		return ErrEmptyTopic
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		logger.Infof("[nats] Publish msg topic:%s payload:%s encode fail:%s", topic, string(msg.Payload), err.Error())
		return err
	}

	logger.Infof("[nats] Publish msg topic:%s payload:%s success", topic, string(msg.Payload))
	err = p.cc.Publish(topic, data)
	if err != nil {
		logger.Infof("[nats] Publish msg topic:%s payload:%s fail:%s", topic, string(msg.Payload), err.Error())
	}
	return err
}

func (p *NatsClient) Subscribe(ctx context.Context, id, topic string, handler client.MessageHandler) error {
	natsFunc := p.natsHandler(handler)
	return p.SubscribeNatsMsg(ctx, id, topic, natsFunc)
}

func (p *NatsClient) SubscribeNatsMsg(ctx context.Context, id, topic string, handler nats.MsgHandler) error {
	if id == "" {
		logger.Infof("[nats] Subscribe id:%s topic:%s fail:%s", id, topic, ErrEmptyID.Error())
		return ErrEmptyID
	}
	if topic == "" {
		logger.Infof("[nats] Subscribe id:%s topic:%s fail:%s", id, topic, ErrEmptyTopic.Error())
		return ErrEmptyTopic
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	s, ok := p.subscriptions[id]
	if ok {
		if _, ok := s.get(topic); ok {
			if err := s.unsubscribe(topic); err != nil {
				logger.Infof("[nats] Subscribe unsubscribe id:%s topic:%s fail:%s", id, topic, ErrEmptyTopic.Error())
				return err
			}
		}
	} else {
		s = subscription{
			queue:         p.queue,
			cc:            p.cc,
			topic2Handler: make(map[string]*subscriptionTopic),
		}
		p.subscriptions[id] = s
	}

	err := s.subscribe(topic, p.queue, handler)
	if err != nil {
		logger.Infof("[nats] Subscribe subscribe id:%s topic:%s fail:%s", id, topic, ErrEmptyTopic.Error())
		return err
	}
	return nil
}

func (p *NatsClient) Unsubscribe(ctx context.Context, id, topic string) error {
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
		return ErrNotSubscribed
	}

	if _, ok := s.get(topic); !ok {
		return ErrNotSubscribed
	}

	if err := s.unsubscribe(topic); err != nil {
		return err
	}
	p.subscriptions[id] = s

	if len(s.topic2Handler) == 0 {
		delete(p.subscriptions, id)
	}
	return nil
}

func (p *NatsClient) Close() error {
	p.cc.GetClient().Close()
	return nil
}

func (p *NatsClient) natsHandler(h client.MessageHandler) nats.MsgHandler {
	return func(m *nats.Msg) {
		msg := &messaging.Message{}
		err := proto.Unmarshal(m.Data, msg)
		if err != nil {
			logger.Infof("[nats] clientId:%s read msg topic:%s payload:%s transform fail:%s", p.clientId, m.Subject, string(m.Data), err.Error())
			return
		}

		logger.Infof("[nats] clientId:%s read msg topic:%s payload:%s success", p.clientId, m.Subject, string(msg.Payload))
		if err := h.Handle(msg); err != nil {
			logger.Warn(fmt.Sprintf("Failed to handle engine message: %s", err))
		}
	}
}
