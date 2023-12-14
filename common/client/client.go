package client

import (
	"context"

	"github.com/visonlv/iot-engine/common/proto/messaging"
)

type MessageHandler interface {
	Handle(msg *messaging.Message) error
	Cancel() error
}

type MessageCommonHandler func(msg *messaging.Message) error

func (h MessageCommonHandler) Handle(msg *messaging.Message) error {
	return h(msg)
}

func (h MessageCommonHandler) Cancel() error {
	return nil
}

type Pub interface {
	Publish(ctx context.Context, topic string, msg *messaging.Message) error
	Close() error
}

type Sub interface {
	Subscribe(ctx context.Context, id, topic string, handler MessageHandler) error
	Unsubscribe(ctx context.Context, id, topic string) error
	Close() error
}

type Proxy interface {
	Pub
	Sub
}
