// Code generated by protoc-gen-vkit.
// versions:
// - protoc-gen-vkit v1.0.0

package handler

import (
	context "context"
	"encoding/json"
	"fmt"

	"github.com/visonlv/go-vkit/errorsx"
	"github.com/visonlv/go-vkit/utilsx"
	"github.com/visonlv/iot-engine/common/define"
	"github.com/visonlv/iot-engine/thing/handler/product"
	"github.com/visonlv/iot-engine/thing/model"
	pb "github.com/visonlv/iot-engine/thing/proto"
	"gorm.io/gorm"
)

type ProductModelService struct {
}

func (the *ProductModelService) Add(ctx context.Context, req *pb.ProductModelAddReq, resp *pb.ProductModelAddResp) error {
	p, err := model.ProductGet(nil, req.ProductId)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("产品查询失败:%s", err.Error())
		return nil
	}
	if p.Transform != string(define.TransformModel) {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = "该产品不支持物模型"
		return nil
	}

	if define.IsSysModelCode(req.Code) {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("%s 为物模型关键词，不可添加", req.Code)
		return nil
	}

	object, err := define.ParseModelItem(req.Name, req.Code, define.ModelType(req.Type), req.ModelDef)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = err.Error()
		return nil
	}

	saveByte, _ := json.Marshal(object)

	// 去重判断
	pm, err := model.ProductModelGetByProductIdAndCodeAndType(nil, req.ProductId, req.Code, req.Type)
	if err != nil && err != gorm.ErrRecordNotFound {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = err.Error()
		return nil
	}
	//新增
	if pm.Id != "" {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("代码已经存在，不可重复添加:%v", req.Code)
		return nil
	}

	m := &model.ProductModelModel{
		Name:      req.Name,
		Code:      req.Code,
		Type:      req.Type,
		Desc:      req.Desc,
		ProductId: req.ProductId,
		ModelDef:  string(saveByte),
		IsSys:     0,
	}
	err = model.ProductModelAdd(nil, m)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("新增物模型项失败:%s", err.Error())
		return nil
	}

	err = product.SyncOneProduct(p)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("同步物模型失败:%s", err.Error())
		return nil
	}
	resp.Id = m.Id
	return nil
}

func (the *ProductModelService) Del(ctx context.Context, req *pb.ProductModelDelReq, resp *pb.ProductModelDelResp) error {
	m, err := model.ProductModelGet(nil, req.Id)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("获取物模型项失败:%s", err.Error())
		return nil
	}

	if m.IsSys == 1 {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = "系统配置不可操作"
		return nil
	}

	p, err := model.ProductGet(nil, m.ProductId)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("产品查询失败:%s", err.Error())
		return nil
	}

	if p.Transform != string(define.TransformModel) {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = "该产品不支持物模型"
		return nil
	}

	m.IsDelete = 1
	err = model.ProductModelDel(nil, req.Id)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("删除物模型项失败:%s", err.Error())
		return nil
	}

	err = product.SyncOneProduct(p)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("同步物模型失败:%s", err.Error())
		return nil
	}
	resp.Id = req.Id
	return nil
}

func (the *ProductModelService) Update(ctx context.Context, req *pb.ProductModelUpdateReq, resp *pb.ProductModelUpdateResp) error {
	m, err := model.ProductModelGet(nil, req.Id)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("获取物模型项失败:%s", err.Error())
		return nil
	}

	if m.IsSys == 1 {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = "系统配置不可操作"
		return nil
	}

	p, err := model.ProductGet(nil, m.ProductId)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("产品查询失败:%s", err.Error())
		return nil
	}

	if p.Transform != string(define.TransformModel) {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = "该产品不支持物模型"
		return nil
	}

	if define.IsSysModelCode(req.Code) {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("%s 为物模型关键词，不可使用", req.Code)
		return nil
	}

	object, err := define.ParseModelItem(req.Name, req.Code, define.ModelType(req.Type), req.ModelDef)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = err.Error()
		return nil
	}
	saveByte, _ := json.Marshal(object)

	// 去重判断
	pm, err := model.ProductModelGetByProductIdAndCodeAndType(nil, m.ProductId, req.Code, req.Type)
	if err != nil && err != gorm.ErrRecordNotFound {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = err.Error()
		return nil
	}
	//更新
	if pm.Id != "" && pm.Id != req.Id {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("代码已经存在，更新失败:%v", req.Code)
		return nil
	}

	m.Name = req.Name
	m.Code = req.Code
	m.Desc = req.Desc
	m.ModelDef = string(saveByte)
	err = model.ProductModelUpdate(nil, m)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("修改物模型项失败:%s", err.Error())
		return nil
	}

	err = product.SyncOneProduct(p)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("同步物模型失败:%s", err.Error())
		return nil
	}
	resp.Id = m.Id
	return nil
}

func (the *ProductModelService) Get(ctx context.Context, req *pb.ProductModelGetReq, resp *pb.ProductModelGetResp) error {
	m, err := model.ProductModelGet(nil, req.Id)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("获取物模型项失败:%s", err.Error())
		return nil
	}

	itemRet := &pb.ProductModel{}
	utilsx.DeepCopy(m, itemRet)
	itemRet.CreateTime = m.CreatedAt.UnixMilli()
	resp.Item = itemRet
	return nil
}

func (the *ProductModelService) List(ctx context.Context, req *pb.ProductModelListReq, resp *pb.ProductModelListResp) error {
	list, err := model.ProductModelList(nil, req.ProductId, req.Name, req.Code, req.Type)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("获取列表失败:%s", err.Error())
		return nil
	}

	listRet := make([]*pb.ProductModel, 0)
	for _, m := range list {
		itemRet := &pb.ProductModel{}
		utilsx.DeepCopy(m, itemRet)
		itemRet.CreateTime = m.CreatedAt.UnixMilli()
		listRet = append(listRet, itemRet)
	}
	resp.Items = listRet
	return nil
}

func (the *ProductModelService) Page(ctx context.Context, req *pb.ProductModelPageReq, resp *pb.ProductModelPageResp) error {
	list, total, err := model.ProductModelPage(nil, req.PageIndex, req.PageSize, req.ProductId, req.Name, req.Code, req.Type)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("获取分页失败:%s", err.Error())
		return nil
	}

	listRet := make([]*pb.ProductModel, 0)
	for _, m := range list {
		itemRet := &pb.ProductModel{}
		utilsx.DeepCopy(m, itemRet)
		itemRet.CreateTime = m.CreatedAt.UnixMilli()
		listRet = append(listRet, itemRet)
	}
	resp.Items = listRet
	resp.Total = total
	return nil
}
