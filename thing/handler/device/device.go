package device

import (
	"context"
	"fmt"

	shadowpb "github.com/visonlv/iot-engine/shadow/proto"
	"github.com/visonlv/iot-engine/thing/app"
)

func GetDevicesProperties(sns []string, codes []string) (map[string]map[string]*shadowpb.ForwardingPropertyItem, error) {
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
