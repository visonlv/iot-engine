package mqtt

import (
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type subscription struct {
	cc            mqtt.Client
	topic2Handler map[string]mqtt.MessageHandler
}

func (s *subscription) unsubscribe(topic string, timeout time.Duration) error {
	token := s.cc.Unsubscribe(topic)
	if token.Error() != nil {
		return token.Error()
	}

	if ok := token.WaitTimeout(timeout); !ok {
		return ErrUnsubscribeTimeout
	}

	if ok := s.delete(topic); !ok {
		return ErrUnsubscribeDeleteTopic
	}
	return token.Error()
}

func (s *subscription) subscribe(topic string, qos int, callback mqtt.MessageHandler, timeout time.Duration) error {
	token := s.cc.Subscribe(topic, QOS, callback)
	if token.Error() != nil {
		return token.Error()
	}
	if ok := token.WaitTimeout(timeout); !ok {
		return ErrSubscribeTimeout
	}
	err := token.Error()
	if err != nil {
		return err
	}
	s.topic2Handler[topic] = callback
	return nil
}

func (s *subscription) get(topic string) (mqtt.MessageHandler, bool) {
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
