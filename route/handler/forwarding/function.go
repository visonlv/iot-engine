package forwarding

import (
	"context"
	"fmt"

	grouppb "github.com/visonlv/iot-engine/group/proto"
	shadowpb "github.com/visonlv/iot-engine/shadow/proto"
)

var _p *Forwarding

func Start() error {
	_p = newForwarding()
	return nil
}

func ReloadClient(data *grouppb.CategoryHeartBeatResp) error {
	return _p.reloadClient(data)
}

func GetClient(sn string) (*shadowpb.ForwardingServiceClient, error) {
	return _p.getClient(sn)
}

func Properties(ctx context.Context, req *shadowpb.ForwardingPropertiesReq, resp *shadowpb.ForwardingPropertiesResp) error {
	return _p.properties(ctx, req, resp)
}

func Watch(ctx context.Context, req *shadowpb.ForwardingWatchReq, s *shadowpb.ForwardingService_WatchServer) error {
	return _p.watch(ctx, req, s)
}

func getItemKey(item *grouppb.CategoryNodeItem) string {
	return fmt.Sprintf("%s:%s", item.Ip, item.Port)
}
