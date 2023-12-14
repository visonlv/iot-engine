package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/visonlv/go-vkit/errorsx/neterrors"
	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/iot-engine/common/client"
	"github.com/visonlv/iot-engine/common/proto/messaging"
)

type httpHandler struct {
	pub client.Pub
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			errorStr := fmt.Sprintf("panic recovered:%v ", err)
			logger.Errorf(errorStr)
			logger.Error(string(debug.Stack()))
			errorResponse(w, r, neterrors.InternalServerError(errorStr))
		}
	}()

	method := strings.ToUpper(r.Method)
	if method == "OPTIONS" {
		return
	}

	if method != "POST" {
		errorStr := fmt.Sprintf("req method:%s only support, url:%s", method, r.RequestURI)
		errorResponse(w, r, neterrors.BadRequest(errorStr))
		return
	}

	topic := r.RequestURI
	body, err := requestPayload(r)
	if err != nil {
		errorResponse(w, r, neterrors.BadRequest(err.Error()))
		return
	}

	pk, sn, _, _, _, isUp, err := client.DecodeMqttTopic(topic)
	if err != nil {
		errorResponse(w, r, neterrors.BadRequest(err.Error()))
		return
	}

	if !isUp {
		errorResponse(w, r, neterrors.BadRequest("need up msg"))
		return
	}

	msg := &messaging.Message{
		Pk:       pk,
		Sn:       sn,
		Topic:    topic,
		Protocol: "http",
		Payload:  body,
		Created:  time.Hour.Nanoseconds(),
	}

	newTopic := client.DeviceClientToNatsUpTopic(msg.Sn)
	if err := h.pub.Publish(context.Background(), newTopic, msg); err != nil {
		errorResponse(w, r, neterrors.BadRequest("pub err:%s", err.Error()))
		return
	}

	okBody := []byte("{\"code\":0,\"msg\":\"\"}")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Header().Set("Content-Length", strconv.Itoa(len(okBody)))
	_, err = w.Write(okBody)
	if err != nil {
		logger.Errorf("response fail url:%v respBytes:%s", r.RequestURI, string(okBody))
	}
}

func errorResponse(w http.ResponseWriter, r *http.Request, _err error) {
	var netErr *neterrors.NetError
	switch verr := _err.(type) {
	case *neterrors.NetError:
		netErr = verr
	default:
		netErr = &neterrors.NetError{
			Msg:    "系统错误",
			Code:   -1,
			Status: 400,
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(int(netErr.Status))

	paramJson, err := json.Marshal(*netErr)
	if err != nil {
		logger.Errorf("encode json fail url:%s", r.RequestURI)
		return
	}

	paramStr := string(paramJson)
	w.Header().Set("Content-Length", strconv.Itoa(len(paramJson)))
	logger.Errorf("with error ret:%s url:%s", paramStr, r.RequestURI)
	fmt.Fprintln(w, paramStr)
}

func requestPayload(r *http.Request) (bytes []byte, err error) {
	closeBody := func(body io.ReadCloser) {
		if e := body.Close(); e != nil {
			err = errors.New(" body close failed")
			return
		}
	}

	ct := r.Header.Get("Content-Type")
	switch {
	case strings.Contains(ct, "application/json"):
		defer closeBody(r.Body)
		bytes, err = io.ReadAll(r.Body)
		return
	case strings.Contains(ct, "application/x-www-form-urlencoded"):
		r.ParseForm()
		vals := make(map[string]string)
		for k, v := range r.Form {
			vals[k] = strings.Join(v, ",")
		}
		b, err := json.Marshal(vals)
		return b, err
	default:
		err = fmt.Errorf(" not support contentType:%s", ct)
		return
	}
}

func newHandler(pub client.Pub) http.Handler {
	return &httpHandler{pub: pub}
}
