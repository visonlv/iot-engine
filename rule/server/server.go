package server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/visonlv/go-vkit/grpcserver"
	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/iot-engine/rule/app"
	"github.com/visonlv/iot-engine/rule/handler"
)

func logFunc(f grpcserver.HandlerFunc) grpcserver.HandlerFunc {
	return func(ctx context.Context, req *grpcserver.GrpcRequest, resp interface{}) error {
		startTime := time.Now()
		err := f(ctx, req, resp)
		costTime := time.Since(startTime)
		bodyObj := req.Payload()
		body, _ := json.Marshal(bodyObj)
		var logText string
		if err != nil {
			logText = fmt.Sprintf("fail cost:[%v] url:[%s] req:[%v] resp:[%v]", costTime.Milliseconds(), req.Method(), string(body), err.Error())
		} else {
			body2, _ := json.Marshal(resp)
			logText = fmt.Sprintf("success cost:[%v] url:[%s] req:[%v] resp:[%v]", costTime.Milliseconds(), req.Method(), string(body), string(body2))
		}
		logger.Infof(logText)
		return err
	}
}

func Start() {
	// 启动grpc服务
	addr := fmt.Sprintf(":%d", app.Cfg.Server.GrpcPort)
	svr := grpcserver.NewServer(
		grpcserver.Name("shadow"),
		grpcserver.GrpcAddr(addr),
		grpcserver.GrpcWrapHandler(logFunc),
	)

	err := svr.RegisterApiEndpoint(handler.GetList(), handler.GetApiEndpoint())
	if err != nil {
		logger.Errorf("[main] RegisterApiEndpoint fail %s", err)
		panic(err)
	}
	svr.Run()
}
