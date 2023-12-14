package server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/visonlv/iot-engine/auth/app"
	"github.com/visonlv/iot-engine/auth/handler"
	"github.com/visonlv/iot-engine/auth/handler/auth"

	"github.com/visonlv/go-vkit/gate"
	"github.com/visonlv/go-vkit/grpcserver"
	"github.com/visonlv/go-vkit/logger"
)

func Start() {
	//初始化权限数据
	auth.AuthObj.Start()

	// 启动grpc服务
	addr := fmt.Sprintf(":%d", app.Cfg.Server.Port)
	svr := grpcserver.NewServer(
		grpcserver.Name("open-cms"),
		grpcserver.GrpcAddr(addr),
		grpcserver.GrpcWrapHandler(grpcLogFunc),
	)

	//grpc通信方式
	err := svr.RegisterApiEndpoint(handler.GetList(), handler.GetApiEndpoint())
	if err != nil {
		logger.Errorf("[main] RegisterApiEndpoint fail %s", err)
		panic(err)
	}

	svr.Run()

}

func logFunc(f gate.HandlerFunc) gate.HandlerFunc {
	return func(ctx context.Context, req *gate.HttpRequest, resp *gate.HttpResponse) error {
		startTime := time.Now()
		err := f(ctx, req, resp)
		costTime := time.Since(startTime)
		body, _, _ := req.Read()
		var logText string
		if err != nil {
			logText = fmt.Sprintf("fail cost:[%v] url:[%v] req:[%v] resp:[%v]", costTime.Milliseconds(), req.Uri(), string(body), err.Error())
		} else {
			logText = fmt.Sprintf("success cost:[%v] url:[%v] req:[%v] resp:[%v]", costTime.Milliseconds(), req.Uri(), string(body), string(resp.Content()))
		}
		logger.Infof(logText)
		return err
	}
}

func grpcLogFunc(f grpcserver.HandlerFunc) grpcserver.HandlerFunc {
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
