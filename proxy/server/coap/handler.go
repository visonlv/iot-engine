package coap

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	coapmessage "github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/message/pool"
	"github.com/plgd-dev/go-coap/v3/mux"
	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/iot-engine/common/client"
	"github.com/visonlv/iot-engine/common/proto/messaging"
)

const (
	startObserve = 0
)

var (
	errMalformedSubtopic = errors.New("malformed subtopic")
	errBadOptions        = errors.New("bad options")
	errNotFound          = errors.New("entity not found")
)

type coapHandler struct {
	pub client.Pub
}

func newHandler(pub client.Pub) mux.HandlerFunc {
	h := &coapHandler{pub: pub}
	return func(w mux.ResponseWriter, r *mux.Message) {
		h.Handler(w, r)
	}
}

func sendResp(w mux.ResponseWriter, resp *pool.Message) {
	if err := w.Conn().WriteMessage(resp); err != nil {
		logger.Info(fmt.Sprintf("Can't set response: %s", err))
	}
}

func (h *coapHandler) Handler(w mux.ResponseWriter, r *mux.Message) {
	cc := w.Conn()
	m := cc.AcquireMessage(cc.Context())
	defer cc.ReleaseMessage(m)

	m.SetCode(codes.Content)
	m.SetToken(r.Token())
	m.SetBody(bytes.NewReader([]byte("{}")))
	m.SetContentFormat(coapmessage.AppJSON)
	defer sendResp(w, m)

	msg, err := decodeMessage(r, r.Code())
	if err != nil {
		logger.Info(fmt.Sprintf("Error decoding message: %s", err))
		m.SetCode(codes.BadRequest)
		return
	}

	if r.Code() == codes.GET {
		err = handleGet(m.Context(), r, w.Conn(), msg)
	} else if r.Code() == codes.POST {
		m.SetCode(codes.Created)
		err = h.pub.Publish(context.Background(), msg.Topic, msg)
		m.SetBody(bytes.NewReader([]byte("{\"msg\":\"post\"}")))
	} else {
		err = errNotFound
	}

	if err != nil {
		switch {
		case err == errBadOptions:
			m.SetCode(codes.BadOption)
		case err == errNotFound:
			m.SetCode(codes.NotFound)
		default:
			m.SetCode(codes.InternalServerError)
		}
	}
}

func getPath(opts coapmessage.Options) string {
	path, err := opts.Path()
	if err != nil {
		logger.Infof("cannot get path: %v", err)
		return ""
	}
	return path
}

func decodeMessage(r *mux.Message, code codes.Code) (*messaging.Message, error) {
	if r.Options() == nil {
		return &messaging.Message{}, errBadOptions
	}
	path, err := r.Options().Path()
	if err != nil {
		return &messaging.Message{}, err
	}

	pk, sn, _, _, _, isUp, err := client.DecodeMqttTopic(path)
	if err != nil {
		return &messaging.Message{}, err
	}
	if !isUp {
		return &messaging.Message{}, fmt.Errorf("Failed to convert topic:%s not support up msg", path)
	}

	ret := &messaging.Message{
		Pk:       pk,
		Sn:       sn,
		Topic:    path,
		Protocol: "coap",
		Created:  time.Now().UnixNano(),
	}

	if r.Body() != nil && code == codes.POST {
		logger.Infof("read body")
		buff, err := io.ReadAll(r.Body())
		if err != nil {
			return ret, err
		}
		ret.Payload = buff
	}
	return ret, nil
}

func handleGet(ctx context.Context, r *mux.Message, c mux.Conn, msg *messaging.Message) error {
	var obs uint32
	obs, err := r.Options().Observe()
	if err != nil {
		logger.Infof("Error reading observe option: %s", err)
		return errBadOptions
	}

	logger.Infof("handleGet obs:%d token:%s", obs, r.Token())
	if obs == startObserve {
		// TODO
		periodicTransmitter(c, r.Token())
		return nil
	}
	return nil
}

func periodicTransmitter(cc mux.Conn, token []byte) {
	for obs := int64(10); ; obs++ {
		err := sendResponse(cc, token, obs)
		if err != nil {
			log.Printf("Error on transmitter, stopping: %v", err)
			return
		}
		time.Sleep(time.Second)
	}
}

func sendResponse(cc mux.Conn, token []byte, obs int64) error {
	m := cc.AcquireMessage(cc.Context())
	defer cc.ReleaseMessage(m)
	m.SetCode(codes.Content)
	m.SetToken(token)
	m.SetBody(bytes.NewReader([]byte("{\"code\":0,\"msg\":\"\"}")))
	m.SetContentFormat(coapmessage.AppJSON)
	if obs >= 0 {
		m.SetObserve(uint32(obs))
	}
	return cc.WriteMessage(m)
}
