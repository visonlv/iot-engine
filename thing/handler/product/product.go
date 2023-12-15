package product

import (
	"encoding/json"
	"fmt"

	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/go-vkit/utilsx"
	"github.com/visonlv/iot-engine/common/define"
	"github.com/visonlv/iot-engine/thing/app"
	"github.com/visonlv/iot-engine/thing/model"
	pb "github.com/visonlv/iot-engine/thing/proto"
	"google.golang.org/protobuf/proto"
)

func GetProductByPksAsMap(pkMap map[string]string) (map[string]*model.ProductModel, error) {
	if len(pkMap) == 0 {
		return make(map[string]*model.ProductModel), nil
	}

	pks := make([]string, 0)
	for _, v := range pkMap {
		pks = append(pks, v)
	}

	list, err := model.ProductGetInPks(nil, pks)
	if err != nil {
		return nil, err
	}

	resultMap := make(map[string]*model.ProductModel)
	for _, v := range list {
		resultMap[v.Pk] = v
	}

	if len(pks) != len(resultMap) {
		return nil, fmt.Errorf("产品查询结果跟入参数量不一致 入参:%d 结果:%d", len(pks), len(resultMap))
	}

	return resultMap, nil
}

func InitAllProduct() {
	list, err := model.ProductList(nil, "", "", "", "")
	if err != nil {
		panic(err)
	}
	for _, m := range list {
		err := SyncOneProduct(m)
		if err != nil {
			panic(err)
		}
	}
}

func LoadThingDef(m *model.ProductModel) (*define.ThingInfo, error) {
	list, err := model.ProductModelList(nil, m.Id, "", "", "")
	if err != nil {
		return nil, err
	}
	thingInfo := &define.ThingInfo{
		Properties: make([]*define.Property, 0),
		Events:     make([]*define.Event, 0),
		Services:   make([]*define.Service, 0),
	}

	for _, v := range list {
		var btype define.ModelType = define.ModelType(v.Type)
		if btype == define.ModelTypeProperty {
			property := &define.Property{}
			err := json.Unmarshal([]byte(v.ModelDef), property)
			if err != nil {
				return nil, err
			}
			thingInfo.Properties = append(thingInfo.Properties, property)
		} else if btype == define.ModelTypeEvent {
			event := &define.Event{}
			err := json.Unmarshal([]byte(v.ModelDef), event)
			if err != nil {
				return nil, err
			}
			thingInfo.Events = append(thingInfo.Events, event)
		} else if btype == define.ModelTypeService {
			service := &define.Service{}
			err := json.Unmarshal([]byte(v.ModelDef), service)
			if err != nil {
				return nil, err
			}
			thingInfo.Services = append(thingInfo.Services, service)
		}
	}

	return thingInfo, nil
}

func SyncOneProduct(m *model.ProductModel) error {
	thingInfo, err := LoadThingDef(m)
	if err != nil {
		return err
	}
	bb, _ := json.Marshal(thingInfo)
	// 加载所有属性
	pbProduct := &pb.Product{}
	utilsx.DeepCopy(m, pbProduct)
	pbProduct.ThingDef = string(bb)

	cc, _ := proto.Marshal(pbProduct)
	err = app.Nats.Publish(define.SysTopicProductUpdate, cc)
	if err != nil {
		logger.Errorf("产品更新 pk:%s 发送nats失败 %s", pbProduct.Pk, err.Error())
	} else {
		logger.Infof("产品更新 pk:%s 发送nats成功", pbProduct.Pk)
	}
	return err
}

func InitProductModel(productId string) error {
	thingInfo := DefaultThingDef()
	return UpdateByThingDef(productId, thingInfo)
}

func MergeToDefaultThingDef(thingDef *define.ThingInfo) *define.ThingInfo {
	sourceDef := DefaultThingDef()
	existProperty := make(map[string]string, 0)
	existEvent := make(map[string]string, 0)
	existService := make(map[string]string, 0)
	for _, v := range sourceDef.Properties {
		existProperty[v.Code] = v.Code
	}
	for _, v := range sourceDef.Events {
		existEvent[v.Code] = v.Code
	}
	for _, v := range sourceDef.Services {
		existService[v.Code] = v.Code
	}

	for _, v := range thingDef.Properties {
		if _, ok := existProperty[v.Code]; !ok {
			sourceDef.Properties = append(sourceDef.Properties, v)
		}
	}
	for _, v := range thingDef.Events {
		if _, ok := existEvent[v.Code]; !ok {
			sourceDef.Events = append(sourceDef.Events, v)
		}
	}
	for _, v := range thingDef.Services {
		if _, ok := existService[v.Code]; !ok {
			sourceDef.Services = append(sourceDef.Services, v)
		}
	}
	return sourceDef
}

func UpdateByThingDef(productId string, thingInfo *define.ThingInfo) error {
	list := make([]*model.ProductModelModel, 0)
	for _, v := range thingInfo.Properties {
		bb, _ := json.Marshal(v)
		item := &model.ProductModelModel{
			ProductId: productId,
			Name:      v.Name,
			Code:      v.Code,
			Type:      string(define.ModelTypeProperty),
			Desc:      v.Desc,
			ModelDef:  string(bb),
			IsSys:     1,
		}
		list = append(list, item)
	}

	for _, v := range thingInfo.Events {
		bb, _ := json.Marshal(v)
		item := &model.ProductModelModel{
			ProductId: productId,
			Name:      v.Name,
			Code:      v.Code,
			Type:      string(define.ModelTypeEvent),
			Desc:      v.Desc,
			ModelDef:  string(bb),
			IsSys:     1,
		}
		list = append(list, item)
	}

	for _, v := range thingInfo.Services {
		bb, _ := json.Marshal(v)
		item := &model.ProductModelModel{
			ProductId: productId,
			Name:      v.Name,
			Code:      v.Code,
			Type:      string(define.ModelTypeService),
			Desc:      v.Desc,
			ModelDef:  string(bb),
			IsSys:     1,
		}
		list = append(list, item)
	}
	return model.ProductModelUpdateByProductId(nil, productId, list)
}

func DefaultThingDef() *define.ThingInfo {
	thingInfo := &define.ThingInfo{
		Properties: make([]*define.Property, 0),
		Events:     make([]*define.Event, 0),
		Services:   make([]*define.Service, 0),
	}
	propertyOnline := &define.Property{
		BaseParamDefine: define.BaseParamDefine{
			Code:        define.PropertyOnline,
			Name:        "在线状态",
			Desc:        "ture 在线 false 离线",
			Required:    true,
			Type:        define.DataTypeBool,
			BoolOptions: &define.BoolOptions{Default: false},
		},
		Mode:       define.PropertyModeRW,
		IsNoRecord: true,
	}

	thingInfo.Properties = append(thingInfo.Properties, propertyOnline)
	return thingInfo
}
