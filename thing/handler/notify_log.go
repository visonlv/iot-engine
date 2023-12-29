// Code generated by protoc-gen-vkit.
// versions:
// - protoc-gen-vkit v1.0.0

package handler

import (
	context "context"
	"fmt"

	"github.com/visonlv/go-vkit/errorsx"
	"github.com/visonlv/go-vkit/utilsx"
	"github.com/visonlv/iot-engine/thing/model"
	pb "github.com/visonlv/iot-engine/thing/proto"
)

type NotifyLogService struct {
}

func (the *NotifyLogService) Del(ctx context.Context, req *pb.NotifyLogDelReq, resp *pb.NotifyLogDelResp) error {
	err := model.NotifyLogDel(nil, req.Id)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("删除通知日志失败:%s", err.Error())
		return nil
	}
	resp.Id = req.Id
	return nil
}

func (the *NotifyLogService) Page(ctx context.Context, req *pb.NotifyLogPageReq, resp *pb.NotifyLogPageResp) error {
	list, total, err := model.NotifyLogPage(nil, req.PageIndex, req.PageSize, req.NotifyConfigId, req.NotifyTemplateId, req.NotifyType)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("获取分页失败:%s", err.Error())
		return nil
	}

	configIdMap := make(map[string]bool)
	templateIdMap := make(map[string]bool)
	for _, m := range list {
		configIdMap[m.NotifyConfigId] = true
		templateIdMap[m.NotifyTemplateId] = true
	}

	configIds := make([]string, 0)
	for v := range configIdMap {
		configIds = append(configIds, v)
	}

	configInfoMap := make(map[string]*pb.NotifyConfig)
	if len(configIds) > 0 {
		list, err := model.NotifyConfigListInIds(nil, configIds)
		if err != nil {
			resp.Code = errorsx.FAIL.Code
			resp.Msg = fmt.Sprintf("获取通知配置列表失败:%s", err.Error())
			return nil
		}

		for _, v := range list {
			itemRet := &pb.NotifyConfig{}
			utilsx.DeepCopy(v, itemRet)
			itemRet.CreateTime = v.CreatedAt.UnixMilli()
			configInfoMap[v.Id] = itemRet
		}
	}

	templateIds := make([]string, 0)
	for v := range templateIdMap {
		templateIds = append(templateIds, v)
	}

	templateInfoMap := make(map[string]*pb.NotifyTemplate)
	if len(templateIds) > 0 {
		list, err := model.NotifyTemplateListInIds(nil, templateIds)
		if err != nil {
			resp.Code = errorsx.FAIL.Code
			resp.Msg = fmt.Sprintf("获取通知配置列表失败:%s", err.Error())
			return nil
		}

		for _, v := range list {
			itemRet := &pb.NotifyTemplate{}
			utilsx.DeepCopy(v, itemRet)
			itemRet.CreateTime = v.CreatedAt.UnixMilli()
			templateInfoMap[v.Id] = itemRet
		}
	}

	listRet := make([]*pb.NotifyLog, 0)
	for _, m := range list {
		itemRet := &pb.NotifyLog{}
		utilsx.DeepCopy(m, itemRet)
		itemRet.CreateTime = m.CreatedAt.UnixMilli()

		if info, ok := templateInfoMap[m.NotifyTemplateId]; ok {
			itemRet.NotifyTemplate = info
		}

		if info, ok := configInfoMap[m.NotifyConfigId]; ok {
			itemRet.NotifyConfig = info
		}
		listRet = append(listRet, itemRet)
	}

	resp.Items = listRet
	resp.Total = total
	return nil
}
