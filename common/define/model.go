package define

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cast"
)

// 数据类型
type DataType string

const (
	DataTypeBool   DataType = "bool"
	DataTypeInt    DataType = "int"
	DataTypeString DataType = "string"
	DataTypeFloat  DataType = "float"
	DataTypeArray  DataType = "array"
	DataTypeObject DataType = "object"
)

// 属性读写类型: r(只读) rw(可读可写)
type PropertyMode string

const (
	PropertyModeR  PropertyMode = "r"
	PropertyModeRW PropertyMode = "rw"
)

// 事件类型: 信息:info  告警alert  故障:fault
type EventType string

const (
	EventTypeInfo  EventType = "info"
	EventTypeAlert EventType = "alert"
	EventTypeFault EventType = "fault"
)

// 服务的执行方向
type ServiceDir string

const (
	ServiceDirUp   ServiceDir = "up"   //向上调用
	ServiceDirDown ServiceDir = "down" //向下调用
)

type ThingInfo struct {
	Properties []*Property `json:"properties"` //属性
	Events     []*Event    `json:"events"`     //事件
	Services   []*Service  `json:"services"`   //服务

	PropertyMap    map[string]*Property `json:"-"` //属性 内部加速索引
	EventMap       map[string]*Event    `json:"-"` //事件 内部加速索引
	UpServiceMap   map[string]*Service  `json:"-"` //服务 内部加速索引
	DownServiceMap map[string]*Service  `json:"-"` //服务 内部加速索引
}

type IntOptions struct {
	Default int32  `json:"default"` //默认值
	Min     int32  `json:"min"`     //数值最小值
	Max     int32  `json:"max"`     //数值最大值
	Step    int32  `json:"step"`    //步长
	Unit    string `json:"unit"`    //单位
}

type BoolOptions struct {
	Default bool `json:"default"` //默认值
}

type FloatOptions struct {
	Default float32 `json:"default"` //默认值
	Min     float32 `json:"min"`     //数值最小值
	Max     float32 `json:"max"`     //数值最大值
	Step    float32 `json:"step"`    //步长
	Unit    string  `json:"unit"`    //单位
}

type StringOptions struct {
	Min int32 `json:"min"` //最小字节长度
	Max int32 `json:"max"` //最大字节长度
}

type ArrayOptions struct {
	Array []*BaseParamDefine `json:"array"` //数组参数
	Min   int32              `json:"min"`   //最小数组长度
	Max   int32              `json:"max"`   //最大数组长度
}

type ObjectOptions struct {
	Object map[string]*BaseParamDefine `json:"object"` //对象参数
}

type BaseParamDefine struct {
	Code     string   `json:"code"`           //标识符
	Name     string   `json:"name"`           //功能名称
	Desc     string   `json:"desc,omitempty"` //描述
	Required bool     `json:"required"`       //是否必须
	Type     DataType `json:"type"`           //属性类型

	IntOptions    *IntOptions    `json:"int_options,omitempty"`
	BoolOptions   *BoolOptions   `json:"bool_options,omitempty"`
	FloatOptions  *FloatOptions  `json:"float_options,omitempty"`
	StringOptions *StringOptions `json:"string_options,omitempty"`
	ArrayOptions  *ArrayOptions  `json:"array_options,omitempty"`
	ObjectOptions *ObjectOptions `json:"object_options,omitempty"`
}

// 属性定义
type Property struct {
	BaseParamDefine
	Mode        PropertyMode `json:"mode"`          //读写类型:rw(可读可写) r(只读)
	IsUseShadow bool         `json:"is_use_shadow"` //是否使用影子
	IsNoRecord  bool         `json:"is_no_record"`  //是否存储历史
}

// 服务定义
type Service struct {
	Code      string            `json:"code"`           //标识符
	Name      string            `json:"name"`           //功能名称
	Desc      string            `json:"desc,omitempty"` //描述
	Dir       ServiceDir        `json:"dir"`            //调用方向
	Input     []*Param          `json:"input"`          //调用参数
	Output    []*Param          `json:"output"`         //返回参数
	InputMap  map[string]*Param `json:"-"`              //调用参数 内部加速索引
	OutputMap map[string]*Param `json:"-"`              //返回参数 内部加速索引
}

// 事件定义
type Event struct {
	Code     string            `json:"code"`           //标识符
	Name     string            `json:"name"`           //功能名称
	Desc     string            `json:"desc,omitempty"` //描述
	Type     EventType         `json:"type"`           //事件类型: 1:信息:info  2:告警alert  3:故障:fault
	Params   []*Param          `json:"params"`         //参数
	ParamMap map[string]*Param `json:"-"`              //参数 内部加速索引
}

// 参数
type Param struct {
	BaseParamDefine
}

func PropertyCurrentValurAsString(val any) (string, error) {
	if object, ok := val.(map[string]any); ok {
		bb, err := json.Marshal(object)
		return string(bb), err
	}
	if array, ok := val.([]any); ok {
		bb, err := json.Marshal(array)
		return string(bb), err
	}
	return cast.ToStringE(val)
}

func ParseVal(d *BaseParamDefine, code string, val any) (any, error) {
	switch d.Type {
	case DataTypeBool:
		return cast.ToBoolE(val)
	case DataTypeInt:
		newVal, err := cast.ToInt32E(val)
		if err != nil {
			return nil, err
		}
		if d.IntOptions.Min != 0 && d.IntOptions.Max != 0 {
			if newVal > d.IntOptions.Max && newVal < d.IntOptions.Min {
				return nil, fmt.Errorf("code:%s 数据范围限制 cur:%d min:%d max:%d", code, newVal, d.IntOptions.Min, d.IntOptions.Max)
			}
		}
		if d.IntOptions.Step != 0 {
			if d.IntOptions.Step != 0 {
				newVal = newVal / d.IntOptions.Step * d.IntOptions.Step
			}
		}
		return newVal, nil
	case DataTypeFloat:
		newVal, err := cast.ToFloat32E(val)
		if err != nil {
			return nil, err
		}
		if d.FloatOptions.Min != 0 && d.FloatOptions.Max != 0 {
			if newVal > d.FloatOptions.Max && newVal < d.FloatOptions.Min {
				return nil, fmt.Errorf("code:%s 数据范围限制 cur:%f min:%f max:%f", code, newVal, d.FloatOptions.Min, d.FloatOptions.Max)
			}
		}
		if d.FloatOptions.Step != 0 {
			if d.FloatOptions.Step != 0 {
				newVal = newVal / d.FloatOptions.Step * d.FloatOptions.Step
			}
		}
		return newVal, nil
	case DataTypeString:
		newVal, err := cast.ToStringE(val)
		if err != nil {
			return nil, err
		}
		var len int32 = int32(len(newVal))
		if d.StringOptions.Min != 0 && d.StringOptions.Max != 0 {
			if len > d.StringOptions.Max && len < d.StringOptions.Min {
				return nil, fmt.Errorf("code:%s 字节长度限制 cur:%d min:%d max:%d", code, len, d.StringOptions.Min, d.StringOptions.Max)
			}
		}
		return newVal, nil
	case DataTypeObject:
		object, ok := val.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("code:%s 类型错误:%T val:%v", code, val, val)
		}
		newObject := make(map[string]any)
		for k, v := range object {
			def, ok := d.ObjectOptions.Object[k]
			if !ok {
				return nil, fmt.Errorf("model object:%s not define property key:%s", code, k)
			}
			newVal, err := ParseVal(def, k, v)
			if err != nil {
				return nil, err
			}
			newObject[k] = newVal
		}
		return newObject, nil
	case DataTypeArray:
		array, ok := val.([]any)
		if !ok {
			return nil, fmt.Errorf("code:%s 类型错误:%T val:%v", code, val, val)
		}
		if d.ArrayOptions.Max != 0 {
			var size int32 = int32(len(array))
			if size > d.ArrayOptions.Max && size < d.ArrayOptions.Min {
				return nil, fmt.Errorf("code:%s 长度限制 cur:%d min:%d max:%d", code, size, d.ArrayOptions.Min, d.ArrayOptions.Max)
			}
		}

		if len(d.ArrayOptions.Array) <= 0 {
			return nil, fmt.Errorf("model array:%s not define item property", code)
		}

		def := d.ArrayOptions.Array[0]
		newArray := make([]any, 0)
		for _, v := range array {
			newVal, err := ParseVal(def, def.Code, v)
			if err != nil {
				return nil, err
			}
			newArray = append(newArray, newVal)
		}
		return newArray, nil
	}
	return nil, fmt.Errorf("type:%s not support", d.Type)
}

func IsThingDefValid(thingDef string) (*ThingInfo, error) {
	m := &ThingInfo{}
	err := json.Unmarshal([]byte(thingDef), m)
	if err != nil {
		return nil, err
	}

	enterCode := make(map[string]string)
	for _, v := range m.Properties {
		if _, ok := enterCode[v.Code]; ok {
			return nil, fmt.Errorf("property code:%s conflict", v.Code)
		}
		enterCode[v.Code] = v.Code
		if v.Code == "" {
			return nil, fmt.Errorf("property code should set")
		}

		if IsSysModelCode(v.Code) {
			return nil, fmt.Errorf("%s 为物模型关键词，不可使用", v.Code)
		}

		if v.Name == "" {
			return nil, fmt.Errorf("property code:%s name should set", v.Code)
		}

		if v.Mode != PropertyModeR && v.Mode != PropertyModeRW {
			return nil, fmt.Errorf("property code:%s mode:%s err", v.Code, v.Mode)
		}
		err := IsBaseDefineValid(&v.BaseParamDefine, false)
		if err != nil {
			return nil, err
		}
	}

	enterCode = make(map[string]string)
	for _, v := range m.Services {
		if _, ok := enterCode[v.Code]; ok {
			return nil, fmt.Errorf("service code:%s conflict", v.Code)
		}
		enterCode[v.Code] = v.Code
		if v.Code == "" {
			return nil, fmt.Errorf("service code should set")
		}

		if IsSysModelCode(v.Code) {
			return nil, fmt.Errorf("%s 为物模型关键词，不可使用", v.Code)
		}

		if v.Name == "" {
			return nil, fmt.Errorf("service code:%s name should set", v.Code)
		}

		if v.Dir != ServiceDirDown && v.Dir != ServiceDirUp {
			return nil, fmt.Errorf("service code:%s dir should be up or down", v.Code)
		}

		if v.Dir != ServiceDirDown && v.Dir != ServiceDirUp {
			return nil, fmt.Errorf("service code:%s dir should be up or down", v.Code)
		}

		if v.Input == nil {
			return nil, fmt.Errorf("service code:%s input should set", v.Code)
		}

		err := IsThingParamsDefValid(v.Input)
		if err != nil {
			return nil, err
		}

		if v.Output == nil {
			return nil, fmt.Errorf("service code:%s output should set", v.Code)
		}

		err = IsThingParamsDefValid(v.Output)
		if err != nil {
			return nil, err
		}
	}

	enterCode = make(map[string]string)
	for _, v := range m.Events {
		if _, ok := enterCode[v.Code]; ok {
			return nil, fmt.Errorf("event code:%s conflict", v.Code)
		}
		enterCode[v.Code] = v.Code
		if v.Code == "" {
			return nil, fmt.Errorf("event code should set")
		}

		if IsSysModelCode(v.Code) {
			return nil, fmt.Errorf("%s 为物模型关键词，不可使用", v.Code)
		}

		if v.Name == "" {
			return nil, fmt.Errorf("event code:%s name should set", v.Code)
		}

		if v.Type != EventTypeInfo && v.Type != EventTypeAlert && v.Type != EventTypeFault {
			return nil, fmt.Errorf("event code:%s type:%s not support", v.Code, v.Type)
		}

		if v.Params == nil {
			return nil, fmt.Errorf("event code:%s Params should set", v.Code)
		}

		err := IsThingParamsDefValid(v.Params)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

func IsThingParamsDefValid(params []*Param) error {
	enterCode := make(map[string]string)
	for _, v := range params {
		if _, ok := enterCode[v.Code]; ok {
			return fmt.Errorf("param code:%s conflict", v.Code)
		}
		err := IsBaseDefineValid(&v.BaseParamDefine, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func ThingDef2Info(thingDef string) (*ThingInfo, error) {
	m := &ThingInfo{}
	err := json.Unmarshal([]byte(thingDef), m)
	if err != nil {
		return nil, err
	}

	m.PropertyMap = make(map[string]*Property)
	for _, v := range m.Properties {
		m.PropertyMap[v.Code] = v
	}

	m.EventMap = make(map[string]*Event)
	for _, v := range m.Events {
		m.EventMap[v.Code] = v
		v.ParamMap = make(map[string]*Param)
		for _, v1 := range v.Params {
			v.ParamMap[v1.Code] = v1
		}
	}

	m.UpServiceMap = make(map[string]*Service)
	m.DownServiceMap = make(map[string]*Service)
	for _, v := range m.Services {
		if v.Dir == ServiceDirUp {
			m.UpServiceMap[v.Code] = v
		} else {
			m.DownServiceMap[v.Code] = v
		}
		v.InputMap = make(map[string]*Param)
		for _, v1 := range v.Input {
			v.InputMap[v1.Code] = v1
		}

		v.OutputMap = make(map[string]*Param)
		for _, v1 := range v.Output {
			v.OutputMap[v1.Code] = v1
		}
	}

	return m, nil
}

func IsBaseDefineValid(v *BaseParamDefine, sub bool) error {
	if v.Code == "" {
		return fmt.Errorf("param code should set")
	}

	if v.Name == "" {
		return fmt.Errorf("param code:%s name should set", v.Code)
	}

	if sub {
		if v.Type != DataTypeBool &&
			v.Type != DataTypeInt &&
			v.Type != DataTypeString &&
			v.Type != DataTypeFloat {
			return fmt.Errorf("支持一层嵌套")
		}
	}

	switch v.Type {
	case DataTypeBool:
		if v.BoolOptions == nil {
			return fmt.Errorf("bool 类型基础参数必须设置")
		}
	case DataTypeInt:
		if v.IntOptions == nil {
			return fmt.Errorf("int 类型基础参数必须设置")
		}
		options := v.IntOptions
		if options.Max < options.Min {
			return fmt.Errorf("最大值不能小于最小值")
		}
		// if options.Step == 0 {
		// 	return fmt.Errorf("步长必须大于0")
		// }
	case DataTypeFloat:
		if v.FloatOptions == nil {
			return fmt.Errorf("float 类型基础参数必须设置")
		}
		options := v.FloatOptions
		if options.Max < options.Min {
			return fmt.Errorf("最大值不能小于最小值")
		}
		// if options.Step == 0 {
		// 	return fmt.Errorf("步长必须大于0")
		// }
	case DataTypeString:
		if v.StringOptions == nil {
			return fmt.Errorf("string 类型基础参数必须设置")
		}
	case DataTypeObject:
		if v.ObjectOptions == nil {
			return fmt.Errorf("object 类型基础参数必须设置")
		}
		options := v.ObjectOptions
		for _, v := range options.Object {
			err := IsBaseDefineValid(v, true)
			if err != nil {
				return err
			}
		}
	case DataTypeArray:
		if v.ArrayOptions == nil {
			return fmt.Errorf("array 类型基础参数必须设置")
		}
		options := v.ArrayOptions
		if options.Max < options.Min {
			return fmt.Errorf("元素最小不能大于元素最多")
		}
		if options.Max == 0 {
			return fmt.Errorf("最大元素必填且>=1")
		}
		if len(options.Array) != 1 {
			return fmt.Errorf("子属下长度必须为1")
		}

		for _, v := range options.Array {
			err := IsBaseDefineValid(v, true)
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("param code:%s type:%s not supoort", v.Code, v.Type)
	}
	return nil
}

func ParsePropertyModelItem(name, code, modelDef string) (*Property, error) {
	item := &Property{}
	err := json.Unmarshal([]byte(modelDef), item)
	if err != nil {
		return nil, err
	}
	if item.Name != name {
		return nil, fmt.Errorf("名称不一致 :%v", item.Name)
	}

	if item.Code != code {
		return nil, fmt.Errorf("代码不一致 :%v", item.Code)
	}

	if item.Mode != PropertyModeR && item.Mode != PropertyModeRW {
		return nil, fmt.Errorf("权限参数错误 :%v", item.Mode)
	}

	err = IsBaseDefineValid(&item.BaseParamDefine, false)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func ParseEventModelItem(name, code, modelDef string) (*Event, error) {
	item := &Event{}
	err := json.Unmarshal([]byte(modelDef), item)
	if err != nil {
		return nil, err
	}
	if item.Name != name {
		return nil, fmt.Errorf("名称不一致 :%v", item.Name)
	}

	if item.Code != code {
		return nil, fmt.Errorf("代码不一致 :%v", item.Code)
	}

	if item.Type != EventTypeInfo && item.Type != EventTypeAlert && item.Type != EventTypeFault {
		return nil, fmt.Errorf("不支持该事件类型 :%v", item.Type)
	}

	err = IsThingParamsDefValid(item.Params)
	if err != nil {
		return nil, err
	}

	return item, nil
}

func ParseServiceModelItem(name, code, modelDef string) (*Service, error) {
	item := &Service{}
	err := json.Unmarshal([]byte(modelDef), item)
	if err != nil {
		return nil, err
	}
	if item.Name != name {
		return nil, fmt.Errorf("名称不一致 :%v", item.Name)
	}

	if item.Code != code {
		return nil, fmt.Errorf("代码不一致 :%v", item.Code)
	}

	if item.Dir != ServiceDirDown && item.Dir != ServiceDirUp {
		return nil, fmt.Errorf("不支持该服务的执行方向 :%v", item.Dir)
	}

	err = IsThingParamsDefValid(item.Input)
	if err != nil {
		return nil, err
	}

	err = IsThingParamsDefValid(item.Output)
	if err != nil {
		return nil, err
	}

	return item, nil
}

func ParseModelItem(name, code string, btype ModelType, modelDef string) (any, error) {
	if btype == ModelTypeProperty {
		return ParsePropertyModelItem(name, code, modelDef)
	} else if btype == ModelTypeEvent {
		return ParseEventModelItem(name, code, modelDef)
	} else if btype == ModelTypeService {
		return ParseServiceModelItem(name, code, modelDef)
	}
	return nil, fmt.Errorf("不支持该类型 :%v", btype)
}

func GetPropertyDefaulValue(def *Property) (string, error) {
	if DataType(def.Type) == DataTypeBool {
		return cast.ToStringE(def.BoolOptions.Default)
	} else if DataType(def.Type) == DataTypeInt {
		return cast.ToStringE(def.IntOptions.Default)
	} else if DataType(def.Type) == DataTypeString {
		return "", nil
	} else if DataType(def.Type) == DataTypeFloat {
		return cast.ToStringE(def.FloatOptions.Default)
	} else if DataType(def.Type) == DataTypeArray {
		return "[]", nil
	} else if DataType(def.Type) == DataTypeObject {
		return "{}", nil
	}
	return "", fmt.Errorf("不支持该类型 :%v", def.Type)
}
