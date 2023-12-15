package device

import (
	"context"
	"fmt"

	shadowpb "github.com/visonlv/iot-engine/shadow/proto"
	"github.com/visonlv/iot-engine/thing/app"
	"github.com/visonlv/iot-engine/thing/model"
)

func GetDeviceByIdsAsMap(idMap map[string]string) (map[string]*model.DeviceModel, error) {
	if len(idMap) == 0 {
		return make(map[string]*model.DeviceModel), nil
	}

	sns := make([]string, 0)
	for _, v := range idMap {
		sns = append(sns, v)
	}

	list, err := model.DeviceGetInIds(nil, sns)
	if err != nil {
		return nil, err
	}

	resultMap := make(map[string]*model.DeviceModel)
	for _, v := range list {
		resultMap[v.Id] = v
	}

	if len(idMap) != len(resultMap) {
		return nil, fmt.Errorf("设备查询结果跟入参数量不一致 入参:%d 结果:%d", len(idMap), len(resultMap))
	}

	return resultMap, nil
}

func GetDevicesProperties(sns []string, codes []string) (map[string]map[string]*shadowpb.ForwardingPropertyItem, error) {
	if len(sns) == 0 {
		return make(map[string]map[string]*shadowpb.ForwardingPropertyItem), nil
	}
	newResp, err := app.Client.ForwardingService.Properties(context.Background(), &shadowpb.ForwardingPropertiesReq{
		Sns:         sns,
		Codes:       codes,
		WithDefault: true,
	})

	if err != nil {
		return nil, err
	}

	if newResp.Code != 0 {
		return nil, fmt.Errorf("请求属性出现错误码 code:%d msg:%s", newResp.Code, newResp.Msg)
	}

	sn2Properties := make(map[string]map[string]*shadowpb.ForwardingPropertyItem)
	for _, v := range newResp.List {
		properties, ok := sn2Properties[v.Sn]
		if !ok {
			properties = make(map[string]*shadowpb.ForwardingPropertyItem)
		}

		for code, v1 := range v.PropertyMap {
			properties[code] = v1
		}
		sn2Properties[v.Sn] = properties

		if codes != nil && len(codes) == 0 && len(properties) != len(codes) {
			return nil, fmt.Errorf("sn%s 代码长度:%d 属性长度:%d 不一致", v.Sn, len(properties), len(codes))
		}
	}

	if len(sns) != len(sn2Properties) {
		return nil, fmt.Errorf("设备sn长度 代码长度:%d 属性长度:%d 不一致", len(sns), len(sn2Properties))
	}

	return sn2Properties, nil
}

func GetDeviceProperties(sn string, codes []string) (map[string]*shadowpb.ForwardingPropertyItem, error) {
	sn2Properties, err := GetDevicesProperties([]string{sn}, codes)
	if err != nil {
		return nil, err
	}
	return sn2Properties[sn], nil
}

func GetDevicesProperty(sns []string, code string) (map[string]*shadowpb.ForwardingPropertyItem, error) {
	sn2Properties, err := GetDevicesProperties(sns, []string{code})
	if err != nil {
		return nil, err
	}
	newMap := make(map[string]*shadowpb.ForwardingPropertyItem)
	for sn, properties := range sn2Properties {
		newMap[sn] = properties[code]
	}
	return newMap, nil
}

func GetDeviceProperty(sn string, code string) (*shadowpb.ForwardingPropertyItem, error) {
	sn2Properties, err := GetDevicesProperties([]string{sn}, []string{code})
	if err != nil {
		return nil, err
	}
	return sn2Properties[sn][code], nil
}
