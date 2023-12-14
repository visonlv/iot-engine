package client

import (
	"errors"
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/visonlv/iot-engine/common/utils"
)

const (
	TOPIC_MAX_LEN    = 7
	TOPIC_UP_VALUE   = "up"
	TOPIC_DOWN_VALUE = "down"
)

var (
	ErrNotTopicNotSupport = errors.New("not support")
)

// iot/up/0/pk/sn/property
func DecodeMqttTopic(topic string) (pk string, sn string, group string, msgType string, code string, isUp bool, err error) {
	if strings.HasPrefix(topic, "/") {
		err = ErrNotTopicNotSupport
		return
	}
	arr := strings.Split(topic, "/")
	if len(arr) < TOPIC_MAX_LEN {
		err = ErrNotTopicNotSupport
		return
	}

	dir := arr[1]
	if dir != TOPIC_UP_VALUE && dir != TOPIC_DOWN_VALUE {
		err = ErrNotTopicNotSupport
		return
	}

	group = arr[2]
	pk = arr[3]
	sn = arr[4]
	msgType = arr[5]
	code = arr[6]
	isUp = dir == TOPIC_UP_VALUE
	return
}

// iot/down/0/pk/sn/property/xx
func DecodeNatsTopic(topic string) (pk string, sn string, group string, msgType string, code string, isUp bool, err error) {
	if strings.HasPrefix(topic, ".") {
		err = ErrNotTopicNotSupport
		return
	}
	arr := strings.Split(topic, ".")
	if len(arr) < TOPIC_MAX_LEN {
		err = ErrNotTopicNotSupport
		return
	}

	dir := arr[1]
	if dir != TOPIC_UP_VALUE && dir != TOPIC_DOWN_VALUE {
		err = ErrNotTopicNotSupport
		return
	}

	group = arr[2]
	pk = arr[3]
	sn = arr[4]
	msgType = arr[5]
	code = arr[6]
	isUp = dir == TOPIC_UP_VALUE
	return
}

func DeviceClientToNatsUpTopic(sn string) string {
	slot := hash(sn) % 100
	c1 := slot / 10
	c2 := slot % 10
	return fmt.Sprintf("global.down.%d.%d", c1, c2)
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func DecodeMqttSystemTopic(topic string) (pk string, sn string, group string, event string, err error) {
	if !strings.HasPrefix(topic, "$SYS") {
		err = ErrNotTopicNotSupport
		return
	}
	arr := strings.Split(topic, "/")
	if len(arr) < 6 {
		err = ErrNotTopicNotSupport
		return
	}

	clientId := arr[4]
	clientIdArr := strings.Split(clientId, "|")
	if len(clientIdArr) != 2 {
		err = ErrNotTopicNotSupport
		return
	}
	pk = clientIdArr[0]
	sn = clientIdArr[1]
	group = fmt.Sprintf("%d", utils.GetGroupId(sn))
	event = arr[5]
	return
}
