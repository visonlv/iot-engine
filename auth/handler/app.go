// Code generated by protoc-gen-vkit.
// versions:
// - protoc-gen-vkit v1.0.0

package handler

import (
	context "context"

	"github.com/visonlv/go-vkit/errorsx"
	"github.com/visonlv/go-vkit/utilsx"
	"github.com/visonlv/iot-engine/auth/model"
	pb "github.com/visonlv/iot-engine/auth/proto"
	"github.com/visonlv/iot-engine/auth/utils"
)

type AppService struct {
}

func (the *AppService) Page(ctx context.Context, req *pb.AppPageReq, resp *pb.AppPageResp) error {
	apps, total, err := model.AppPage(nil, req.PageIndex, req.PageSize, req.Name)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = "查询应用列表失败：" + err.Error()
		return nil
	}
	for _, m := range apps {
		pbItem := &pb.App{}
		utilsx.DeepCopy(m, pbItem)
		pbItem.CreateTime = m.CreatedAt.UnixMilli()
		pbItem.CreateUsername = model.UserNameFromCache(nil, m.CreateUser)
		resp.List = append(resp.List, pbItem)
	}
	resp.Total = total
	resp.Code = errorsx.OK.Code
	return nil
}

func (the *AppService) Add(ctx context.Context, req *pb.AppAddReq, resp *pb.AppAddResp) error {
	_, exist, err := model.AppGetByCode(nil, req.Code)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = "添加应用失败:" + err.Error()
		return nil
	}
	if exist {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = "添加应用失败:应用代码已存在"
		return nil
	}
	item := &model.AppModel{
		Code:       req.Code,
		Name:       req.Name,
		Describe:   req.Describe,
		CreateUser: utils.GetUserIdFromContext(ctx),
	}
	err = model.AppAdd(nil, item)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = "添加应用失败:" + err.Error()
		return nil
	}
	resp.Code = errorsx.OK.Code
	resp.Id = item.Id
	return nil
}

func (the *AppService) Del(ctx context.Context, req *pb.AppDelReq, resp *pb.AppDelResp) error {
	count, err := model.PermissionCountByAppId(nil, req.Id)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = "查询失败:" + err.Error()
		return nil
	}
	if count > 0 {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = "该应用存在绑定的权限"
		return nil
	}

	err = model.AppDel(nil, req.Id)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = "删除应用失败:" + err.Error()
		return nil
	}
	resp.Code = errorsx.OK.Code
	return nil
}

func (the *AppService) Update(ctx context.Context, req *pb.AppUpdateReq, resp *pb.AppUpdateResp) error {
	item, err := model.AppGet(nil, req.Id)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = "更新应用失败:" + err.Error()
		return nil
	}
	if item.Code != req.Code {
		_, exist, err := model.AppGetByCode(nil, req.Code)
		if err != nil {
			resp.Code = errorsx.FAIL.Code
			resp.Msg = "更新应用失败:" + err.Error()
			return nil
		}
		if exist {
			resp.Code = errorsx.FAIL.Code
			resp.Msg = "更新应用失败:应用标识已存在"
			return nil
		}
	}

	item.Name = req.Name
	item.Code = req.Code
	item.Describe = req.Describe
	err = model.AppUpdate(nil, item)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = "更新应用失败:" + err.Error()
		return nil
	}
	resp.Code = errorsx.OK.Code
	return nil
}

func (the *AppService) Get(ctx context.Context, req *pb.AppGetReq, resp *pb.AppGetResp) error {
	item, err := model.AppGet(nil, req.Id)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = "查询应用失败:" + err.Error()
		return nil
	}
	//查询用户
	u, err := model.UserGetFromCache(nil, item.CreateUser)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = "查询应用失败:" + err.Error()
		return nil
	}
	resp.Code = errorsx.OK.Code
	resp.App = &pb.App{
		Id:             item.Id,
		Code:           item.Code,
		Name:           item.Name,
		Describe:       item.Describe,
		CreateUser:     item.CreateUser,
		CreateUsername: u.NickName,
		CreateTime:     item.CreatedAt.UnixMilli(),
	}

	return nil
}
