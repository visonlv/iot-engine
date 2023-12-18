// Code generated by protoc-gen-vkit. DO NOT EDIT.
// versions:
// - protoc-gen-vkit v1.0.0
// - protoc             v3.21.1
// source: rule.proto

package proto

import (
	context "context"
	grpcx "github.com/visonlv/go-vkit/grpcx"
	grpc "google.golang.org/grpc"
)

var _ = new(context.Context)
var _ = new(grpc.CallOption)
var _ = new(grpcx.Client)

type RuleServiceClient struct {
	name string
	cc   grpcx.Client
}

func (c *RuleServiceClient) Add(ctx context.Context, in *RuleAddReq, opts ...grpc.CallOption) (*RuleAddResp, error) {
	out := new(RuleAddResp)
	err := c.cc.Invoke(ctx, c.name, "RuleService.Add", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *RuleServiceClient) Del(ctx context.Context, in *RuleDelReq, opts ...grpc.CallOption) (*RuleDelResp, error) {
	out := new(RuleDelResp)
	err := c.cc.Invoke(ctx, c.name, "RuleService.Del", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *RuleServiceClient) Update(ctx context.Context, in *RuleUpdateReq, opts ...grpc.CallOption) (*RuleUpdateResp, error) {
	out := new(RuleUpdateResp)
	err := c.cc.Invoke(ctx, c.name, "RuleService.Update", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *RuleServiceClient) Get(ctx context.Context, in *RuleGetReq, opts ...grpc.CallOption) (*RuleGetResp, error) {
	out := new(RuleGetResp)
	err := c.cc.Invoke(ctx, c.name, "RuleService.Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *RuleServiceClient) List(ctx context.Context, in *RuleListReq, opts ...grpc.CallOption) (*RuleListResp, error) {
	out := new(RuleListResp)
	err := c.cc.Invoke(ctx, c.name, "RuleService.List", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *RuleServiceClient) Page(ctx context.Context, in *RulePageReq, opts ...grpc.CallOption) (*RulePageResp, error) {
	out := new(RulePageResp)
	err := c.cc.Invoke(ctx, c.name, "RuleService.Page", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func NewRuleServiceClient(name string, cc grpcx.Client) *RuleServiceClient {
	return &RuleServiceClient{name, cc}
}