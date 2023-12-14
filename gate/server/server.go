package server

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/visonlv/go-vkit/errorsx"
	"github.com/visonlv/go-vkit/errorsx/neterrors"
	"github.com/visonlv/go-vkit/gate"
	"github.com/visonlv/go-vkit/grpcclient"
	"github.com/visonlv/go-vkit/logger"
	authpb "github.com/visonlv/iot-engine/auth/proto"
	"github.com/visonlv/iot-engine/gate/app"
)

var (
	whiteList = []string{}
)

func IsWhite(url string) bool {
	for _, v := range whiteList {
		b, err := path.Match(v, url)
		if err != nil {
			logger.Errorf("IsWhite match url:%s err:%s", url, err)
			return false
		}
		if b {
			return true
		}
	}
	return false
}

// 统一鉴权逻辑
func tokenCheckFunc(w http.ResponseWriter, r *http.Request) error {
	// 判断白名单
	if IsWhite(r.RequestURI) {
		return nil
	}

	// 判断token
	tokenStr := r.Header.Get("AuthToken")
	out, err := app.Client.AuthService.APIPermissions(context.Background(), &authpb.APIPermissionsReq{Token: tokenStr, Api: r.RequestURI})
	if err != nil {
		return neterrors.Unauthorized(err.Error())
	}
	if out.Code != errorsx.OK.Code {
		return neterrors.Unauthorized(out.Msg)
	}
	if out.Code == errorsx.FAIL.Code {
		return neterrors.New(out.Msg, "", -1, out.HttpStatus)
	}
	if out.IsWhite {
		return nil
	}
	if !out.Enable {
		return neterrors.Forbidden("资源没有权限!")
	}

	r.Header.Set("user_id", out.UserId)
	r.Header.Set("token", tokenStr)
	r.Header.Set("role_code", out.RoleCodes[0])
	return nil
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

func success(w http.ResponseWriter, r *http.Request, resultBody []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Header().Set("Content-Length", strconv.Itoa(len(resultBody)))
	w.Write(resultBody)
}

func Start() {
	customHandler := gate.NewGrpcHandler(
		gate.HttpGrpcPort(int(app.Cfg.Server.GrpcProxyPort)),
		gate.HttpAuthHandler(tokenCheckFunc),
		gate.HttpWrapHandler(logFunc))

	grpcclient.SetServerName2Addr(app.Cfg.Business.TargetMap)

	http.HandleFunc("/rpc/", func(w http.ResponseWriter, r *http.Request) {
		customHandler.Handle(w, r)
	})

	logger.Infof("[main] Listen port:%d", int(app.Cfg.Server.HttpPort))
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", app.Cfg.Server.Address, int(app.Cfg.Server.HttpPort)), nil)
	if err != nil {
		logger.Errorf("[main] ListenAndServe fail %s", err)
		panic(err)
	}
}
