package define

import (
	"time"
)

type Shadow struct {
	Properties map[string]*ShadowProperty `json:"properties"` //属性
}

func DefaultShadow() *Shadow {
	s := &Shadow{
		Properties: make(map[string]*ShadowProperty),
	}
	s.Properties[PropertyOnline] = &ShadowProperty{
		Current: &CurrentProperty{Value: false, UpdatedTime: time.Now().UnixMilli()},
	}
	return s
}

func (s *Shadow) PropertyEqualTo(key string, value bool) bool {
	if v, ok := s.Properties[key]; ok && v.Current != nil {
		value, ok := v.Current.Value.(bool)
		if !ok {
			return false
		}
		return value
	}
	return false
}

// 属性状态
type ShadowProperty struct {
	Current *CurrentProperty `json:"current,omitempty"` //当前属性
	Desired *DesiredProperty `json:"desired,omitempty"` //期望属性
}

// 属性当前状态
type CurrentProperty struct {
	Value       any   `json:"value"`        //当前属性值
	UpdatedTime int64 `json:"updated_time"` //更新时间
}

// 属性期望状态
type DesiredProperty struct {
	Value        any   `json:"value"`         //期望属性值
	ReceivedTime int64 `json:"received_time"` //接收时间
	Expiration   int32 `json:"expiration"`    //毫秒
}
