// Code generated by protoc-gen-vkit.
// versions:
// - protoc-gen-vkit v1.0.0

package handler

import (
	context "context"
	"fmt"

	"github.com/spf13/cast"
	"github.com/visonlv/go-vkit/errorsx"
	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/go-vkit/utilsx"
	"github.com/visonlv/iot-engine/common/define"
	"github.com/visonlv/iot-engine/common/utils"
	shadowpb "github.com/visonlv/iot-engine/shadow/proto"
	"github.com/visonlv/iot-engine/thing/app"
	"github.com/visonlv/iot-engine/thing/handler/device"
	"github.com/visonlv/iot-engine/thing/handler/product"
	"github.com/visonlv/iot-engine/thing/model"
	pb "github.com/visonlv/iot-engine/thing/proto"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

type DeviceService struct {
}

func (the *DeviceService) Add(ctx context.Context, req *pb.DeviceAddReq, resp *pb.DeviceAddResp) error {
	_, err := model.DeviceGetBySn(nil, req.Sn)
	if err != nil && err != gorm.ErrRecordNotFound {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("获取设备信息错误:%s", err.Error())
		return nil
	}
	if err == nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = "设备Sn已经存在"
		return nil
	}

	p, err := model.ProductGetByPk(nil, req.Pk)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("获取产品信息失败:%s", err.Error())
		return nil
	}

	var PSn string
	//子设备
	if p.Type == string(define.ProductTypeChild) {
		if req.PId == "" {
			resp.Code = errorsx.FAIL.Code
			resp.Msg = "子设备必须指定网关设备"
			return nil
		}

		pd, err := model.DeviceGet(nil, req.PId)
		if err != nil {
			resp.Code = errorsx.FAIL.Code
			resp.Msg = fmt.Sprintf("获取网关设备信息失败:%s", err.Error())
			return nil
		}
		PSn = pd.Sn

		pp, err := model.ProductGetByPk(nil, pd.Pk)
		if err != nil {
			resp.Code = errorsx.FAIL.Code
			resp.Msg = fmt.Sprintf("获取网关产品信息失败:%s", err.Error())
			return nil
		}

		if pp.Type != string(define.ProductTypeGateway) {
			resp.Code = errorsx.FAIL.Code
			resp.Msg = fmt.Sprintf("指定的网关设备不是网关产品:%s", pp.Type)
			return nil
		}
	}

	m := &model.DeviceModel{
		Id:     utilsx.GenUuid(),
		Pk:     req.Pk,
		Name:   req.Name,
		Sn:     req.Sn,
		PId:    req.PId,
		Group:  utils.GetGroupId(req.Sn),
		Secret: req.Secret,
		Desc:   req.Desc,
	}

	newResp, err := app.Client.ShadowService.Add(context.Background(), &shadowpb.ShadowAddReq{
		Id:  m.Id,
		Sn:  req.Sn,
		Pk:  req.Pk,
		PSn: PSn,
		PId: req.PId,
	})

	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("创建影子失败:%s", err.Error())
		return nil
	}

	if newResp.Code != 0 {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("创建影子失败:%d %s", newResp.Code, newResp.Msg)
		return nil
	}

	err = model.DeviceAdd(nil, m)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("添加设备失败:%s", err.Error())
		return nil
	}

	pbDevice := &pb.Device{}
	utilsx.DeepCopy(m, pbDevice)
	cc, _ := proto.Marshal(pbDevice)
	err = app.Nats.Publish(define.SysTopicDeviceAdd, cc)
	if err != nil {
		logger.Errorf("设备添加 sn:%s 发送nats失败 %s", pbDevice.Sn, err.Error())
	} else {
		logger.Infof("设备添加 sn:%s 发送nats成功", pbDevice.Sn)
	}

	resp.Id = m.Id
	return nil
}

func (the *DeviceService) Del(ctx context.Context, req *pb.DeviceDelReq, resp *pb.DeviceDelResp) error {
	m, err := model.DeviceGet(nil, req.Id)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("获取设备失败:%s", err.Error())
		return nil
	}

	newResp, err := app.Client.ShadowService.Del(context.Background(), &shadowpb.ShadowDelReq{
		Id: m.Id,
	})

	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("删除影子失败:%s", err.Error())
		return nil
	}

	if newResp.Code != 0 {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("删除影子失败:%d %s", newResp.Code, newResp.Msg)
		return nil
	}

	m.IsDelete = 1
	err = model.DeviceUpdate(nil, m)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("删除设备失败:%s", err.Error())
		return nil
	}

	pbDevice := &pb.Device{}
	utilsx.DeepCopy(m, pbDevice)
	cc, _ := proto.Marshal(pbDevice)
	err = app.Nats.Publish(define.SysTopicDeviceDel, cc)
	if err != nil {
		logger.Errorf("设备删除 sn:%s 发送nats失败 %s", pbDevice.Sn, err.Error())
	} else {
		logger.Infof("设备删除 sn:%s 发送nats成功", pbDevice.Sn)
	}

	resp.Id = req.Id
	return nil
}

func (the *DeviceService) Update(ctx context.Context, req *pb.DeviceUpdateReq, resp *pb.DeviceUpdateResp) error {
	m, err := model.DeviceGet(nil, req.Id)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("获取设备失败:%s", err.Error())
		return nil
	}

	p, err := model.ProductGetByPk(nil, m.Pk)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("获取产品信息失败:%s", err.Error())
		return nil
	}

	var PSn string
	//子设备
	if p.Type == string(define.ProductTypeChild) {
		if req.PId == "" {
			resp.Code = errorsx.FAIL.Code
			resp.Msg = "子设备必须指定网关设备"
			return nil
		}

		pd, err := model.DeviceGet(nil, req.PId)
		if err != nil {
			resp.Code = errorsx.FAIL.Code
			resp.Msg = fmt.Sprintf("获取网关设备信息失败:%s", err.Error())
			return nil
		}
		PSn = pd.Sn

		pp, err := model.ProductGetByPk(nil, pd.Pk)
		if err != nil {
			resp.Code = errorsx.FAIL.Code
			resp.Msg = fmt.Sprintf("获取网关产品信息失败:%s", err.Error())
			return nil
		}

		if pp.Type != string(define.ProductTypeGateway) {
			resp.Code = errorsx.FAIL.Code
			resp.Msg = fmt.Sprintf("指定的网关设备不是网关产品:%s", pp.Type)
			return nil
		}
	}

	newResp, err := app.Client.ShadowService.Update(context.Background(), &shadowpb.ShadowUpdateReq{
		Id:  m.Id,
		Sn:  m.Sn,
		Pk:  m.Pk,
		PSn: PSn,
		PId: req.PId,
	})

	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("更新影子失败:%s", err.Error())
		return nil
	}

	if newResp.Code != 0 {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("更新影子失败:%d %s", newResp.Code, newResp.Msg)
		return nil
	}

	m.Name = req.Name
	m.Secret = req.Secret
	m.Desc = req.Desc
	m.PId = req.PId
	err = model.DeviceUpdate(nil, m)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("修改设备失败:%s", err.Error())
		return nil
	}

	pbDevice := &pb.Device{}
	utilsx.DeepCopy(m, pbDevice)
	cc, _ := proto.Marshal(pbDevice)
	err = app.Nats.Publish(define.SysTopicDeviceUpdate, cc)
	if err != nil {
		logger.Errorf("设备更新 sn:%s 发送nats失败 %s", pbDevice.Sn, err.Error())
	} else {
		logger.Infof("设备更新 sn:%s 发送nats成功", pbDevice.Sn)
	}

	resp.Id = m.Id
	return nil
}

func (the *DeviceService) Get(ctx context.Context, req *pb.DeviceGetReq, resp *pb.DeviceGetResp) error {
	m, err := model.DeviceGet(nil, req.Id)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("获取设备失败:%s", err.Error())
		return nil
	}

	var pName string
	if m.PId != "" {
		idMap := make(map[string]string)
		idMap[m.PId] = m.PId
		resultMap, err := device.GetDeviceByIdsAsMap(idMap)
		if err != nil {
			resp.Code = errorsx.FAIL.Code
			resp.Msg = err.Error()
			return nil
		}
		pName = resultMap[m.PId].Name
	}

	pkMap := make(map[string]string)
	pkMap[m.Pk] = m.Pk
	resultPkMap, err := product.GetProductByPksAsMap(pkMap)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = err.Error()
		return nil
	}

	itemRet := &pb.Device{}
	utilsx.DeepCopy(m, itemRet)
	itemRet.CreateTime = m.CreatedAt.UnixMilli()
	itemRet.PName = pName
	itemRet.ProductName = resultPkMap[m.Pk].Name
	itemRet.ProductType = resultPkMap[m.Pk].Type
	resp.Item = itemRet
	return nil
}

func (the *DeviceService) List(ctx context.Context, req *pb.DeviceListReq, resp *pb.DeviceListResp) error {
	list, err := model.DeviceList(nil, req.Pk, req.Name, req.Sn, req.PId)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("获取列表失败:%s", err.Error())
		return nil
	}

	idMap := make(map[string]string)
	pkMap := make(map[string]string)
	for _, m := range list {
		pkMap[m.Pk] = m.Pk
		if m.PId != "" {
			idMap[m.PId] = m.PId
		}
	}

	resultMap, err := device.GetDeviceByIdsAsMap(idMap)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = err.Error()
		return nil
	}

	resultPkMap, err := product.GetProductByPksAsMap(pkMap)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = err.Error()
		return nil
	}

	listRet := make([]*pb.Device, 0)
	for _, m := range list {
		itemRet := &pb.Device{}
		utilsx.DeepCopy(m, itemRet)
		itemRet.CreateTime = m.CreatedAt.UnixMilli()
		if pm, ok := resultMap[m.PId]; ok {
			itemRet.PName = pm.Name
		}
		itemRet.ProductName = resultPkMap[m.Pk].Name
		itemRet.ProductType = resultPkMap[m.Pk].Type
		listRet = append(listRet, itemRet)
	}
	resp.Items = listRet
	return nil
}

func (the *DeviceService) Page(ctx context.Context, req *pb.DevicePageReq, resp *pb.DevicePageResp) error {
	list, total, err := model.DevicePage(nil, req.PageIndex, req.PageSize, req.Pk, req.Name, req.Sn, req.PId)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("获取分页失败:%s", err.Error())
		return nil
	}

	deviceSns := make([]string, 0)
	idMap := make(map[string]string)
	pkMap := make(map[string]string)
	for _, m := range list {
		deviceSns = append(deviceSns, m.Sn)
		pkMap[m.Pk] = m.Pk
		if m.PId != "" {
			idMap[m.PId] = m.PId
		}
	}

	resultMap, err := device.GetDeviceByIdsAsMap(idMap)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = err.Error()
		return nil
	}

	resultPkMap, err := product.GetProductByPksAsMap(pkMap)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = err.Error()
		return nil
	}

	sn2property, err := device.GetDevicesProperty(deviceSns, define.PropertyOnline)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("获取在线属性失败:%s", err.Error())
		return nil
	}
	listRet := make([]*pb.Device, 0)
	for _, m := range list {
		itemRet := &pb.Device{}
		utilsx.DeepCopy(m, itemRet)
		itemRet.CreateTime = m.CreatedAt.UnixMilli()
		if p, ok := sn2property[m.Sn]; ok {
			itemRet.Online = p.Value == "true"
		}
		if pm, ok := resultMap[m.PId]; ok {
			itemRet.PName = pm.Name
		}
		itemRet.ProductName = resultPkMap[m.Pk].Name
		itemRet.ProductType = resultPkMap[m.Pk].Type

		listRet = append(listRet, itemRet)
	}

	resp.Items = listRet
	resp.Total = total
	return nil
}

func (the *DeviceService) Auth(ctx context.Context, req *pb.DeviceAuthReq, resp *pb.DeviceAuthResp) error {
	// "result": "allow", // 可选 "allow" | "deny" | "ignore"
	resp.Result = "allow"
	resp.IsSuperuser = true
	return nil
}

func (the *DeviceService) Properties(ctx context.Context, req *pb.DevicePropertiesReq, resp *pb.DevicePropertiesResp) error {
	m, err := model.DeviceGet(nil, req.Id)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("获取设备失败:%s", err.Error())
		return nil
	}

	pm, err := model.ProductGetByPk(nil, m.Pk)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("获取产品失败:%s", err.Error())
		return nil
	}

	shadownMap, err := device.GetDeviceProperties(m.Sn, req.Codes)
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("获取当前属性失败:%s", err.Error())
		return nil
	}

	list, err := model.ProductModelList(nil, pm.Id, "", "", string(define.ModelTypeProperty))
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("获取属性物模型失败:%s", err.Error())
		return nil
	}

	var codeMap map[string]string
	if req.Codes != nil && len(req.Codes) > 0 {
		codeMap = map[string]string{}
		for _, v := range req.Codes {
			codeMap[v] = v
		}
	}

	newList := make([]*pb.DeviceProperty, 0)
	for _, v := range list {
		if codeMap != nil {
			if _, ok := codeMap[v.Code]; !ok {
				continue
			}
		}

		property, err := define.ParsePropertyModelItem(v.Name, v.Code, v.ModelDef)
		if err != nil {
			resp.Code = errorsx.FAIL.Code
			resp.Msg = err.Error()
			return nil
		}
		item := &pb.DeviceProperty{
			Code:       v.Code,
			Name:       v.Name,
			Desc:       v.Desc,
			Type:       string(property.Type),
			Value:      "",
			UpdateTime: 0,
		}
		newList = append(newList, item)

		cur, ok := shadownMap[v.Code]
		if ok {
			item.Value = cur.Value
			item.UpdateTime = cur.UpdateTime
			continue
		}

		if item.Type == string(define.DataTypeBool) {
			item.Value = cast.ToString(property.BoolOptions.Default)
		} else if item.Type == string(define.DataTypeInt) {
			item.Value = cast.ToString(property.IntOptions.Default)
		} else if item.Type == string(define.DataTypeFloat) {
			item.Value = cast.ToString(property.FloatOptions.Default)
		}
	}

	resp.Id = req.Id
	resp.Items = newList
	return nil
}

func (the *DeviceService) ListGateway(ctx context.Context, req *pb.DeviceListGatewayReq, resp *pb.DeviceListGatewayResp) error {
	list, err := model.DeviceGetByProductType(nil, string(define.ProductTypeGateway))
	if err != nil {
		resp.Code = errorsx.FAIL.Code
		resp.Msg = fmt.Sprintf("获取设备列表失败:%s", err.Error())
		return nil
	}
	listRet := make([]*pb.Device, 0)
	for _, m := range list {
		itemRet := &pb.Device{}
		utilsx.DeepCopy(m, itemRet)
		itemRet.CreateTime = m.CreatedAt.UnixMilli()
		listRet = append(listRet, itemRet)
	}

	resp.Items = listRet
	return nil
}
