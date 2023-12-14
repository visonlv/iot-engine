// Code generated by protoc-gen-vkit. DO NOT EDIT.
// versions:
// - protoc-gen-vkit v1.0.0

package handler

import (
	"github.com/visonlv/go-vkit/grpcx"
	
)

func GetList() []interface{} {
	list := make([]interface{}, 0)
	list = append(list, &ForwardingService{})
	list = append(list, &ShadowService{})
	list = append(list, &MsgLogService{})
	
	return list
}

func GetApiEndpoint() []*grpcx.ApiEndpoint {
	return []*grpcx.ApiEndpoint{
		{
			Method:"ForwardingService.Properties",
			Url:"/rpc/shadow/ForwardingService.Properties", 
			ClientStream:false, 
			ServerStream:false,
		},{
			Method:"ForwardingService.Service",
			Url:"/rpc/shadow/ForwardingService.Service", 
			ClientStream:false, 
			ServerStream:false,
		},{
			Method:"ForwardingService.SetProperty",
			Url:"/rpc/shadow/ForwardingService.SetProperty", 
			ClientStream:false, 
			ServerStream:false,
		},{
			Method:"ForwardingService.SetProperties",
			Url:"/rpc/shadow/ForwardingService.SetProperties", 
			ClientStream:false, 
			ServerStream:false,
		},{
			Method:"ForwardingService.ServiceReply",
			Url:"/rpc/shadow/ForwardingService.ServiceReply", 
			ClientStream:false, 
			ServerStream:false,
		},{
			Method:"ForwardingService.PublishMsg",
			Url:"/rpc/shadow/ForwardingService.PublishMsg", 
			ClientStream:false, 
			ServerStream:false,
		},{
			Method:"ForwardingService.Watch",
			Url:"/rpc/shadow/ForwardingService.Watch", 
			ClientStream:false, 
			ServerStream:true,
		},{
			Method:"ShadowService.Add",
			Url:"/rpc/shadow/ShadowService.Add", 
			ClientStream:false, 
			ServerStream:false,
		},{
			Method:"ShadowService.Del",
			Url:"/rpc/shadow/ShadowService.Del", 
			ClientStream:false, 
			ServerStream:false,
		},{
			Method:"ShadowService.Update",
			Url:"/rpc/shadow/ShadowService.Update", 
			ClientStream:false, 
			ServerStream:false,
		},{
			Method:"ShadowService.Get",
			Url:"/rpc/shadow/ShadowService.Get", 
			ClientStream:false, 
			ServerStream:false,
		},{
			Method:"ShadowService.List",
			Url:"/rpc/shadow/ShadowService.List", 
			ClientStream:false, 
			ServerStream:false,
		},{
			Method:"ShadowService.Page",
			Url:"/rpc/shadow/ShadowService.Page", 
			ClientStream:false, 
			ServerStream:false,
		},{
			Method:"MsgLogService.Page",
			Url:"/rpc/shadow/MsgLogService.Page", 
			ClientStream:false, 
			ServerStream:false,
		},
	}
}
