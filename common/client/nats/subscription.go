package nats

import (
	"github.com/nats-io/nats.go"
	"github.com/visonlv/go-vkit/natsx"
)

type subscriptionTopic struct {
	h   nats.MsgHandler
	sub *nats.Subscription
}

type subscription struct {
	queue         string
	cc            *natsx.NatsClient
	topic2Handler map[string]*subscriptionTopic
}

func (s *subscription) unsubscribe(topic string) error {
	topicInfo, ok := s.topic2Handler[topic]
	if !ok {
		return ErrNotSubscribed
	}

	if ok := s.delete(topic); !ok {
		return ErrUnsubscribeDeleteTopic
	}

	err := topicInfo.sub.Unsubscribe()
	if err != nil {
		return err
	}

	return nil
}

func (s *subscription) subscribe(topic, queue string, callback nats.MsgHandler) error {
	var sub *nats.Subscription
	var err error
	if queue != "" {
		sub, err = s.cc.QueueSubscribe(topic, queue, callback)
	} else {
		sub, err = s.cc.Subscribe(topic, callback)
	}
	if err != nil {
		return err
	}
	s.topic2Handler[topic] = &subscriptionTopic{h: callback, sub: sub}
	return nil
}

func (s *subscription) get(topic string) (*subscriptionTopic, bool) {
	h, ok := s.topic2Handler[topic]
	return h, ok
}

func (s *subscription) delete(topic string) bool {
	_, ok := s.topic2Handler[topic]
	if !ok {
		return false
	}
	delete(s.topic2Handler, topic)
	return true
}
