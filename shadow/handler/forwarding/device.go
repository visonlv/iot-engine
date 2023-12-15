package forwarding

import (
	"encoding/json"

	"github.com/visonlv/iot-engine/common/define"
	"github.com/visonlv/iot-engine/common/utils"
	"github.com/visonlv/iot-engine/shadow/model"
)

type Device struct {
	Sn            string
	PSn           string
	Group         int32
	Pk            string
	Shadow        *define.Shadow
	ShadowVersion int64
}

type Devices struct {
	deviceLru *utils.LRU[*Device]
	products  *Products
}

type deviceLoadResult struct {
	d   *Device
	err error
}

type saveShadowItem struct {
	sn      string
	shadow  string
	version int64
}

func newDevices(lruSize int, products *Products) *Devices {
	return &Devices{
		deviceLru: utils.NewLRU[*Device](lruSize),
		products:  products,
	}
}

func (d *Devices) getDeviceAndProduct(sn string) (*Device, *Product, error) {
	info, ok := d.deviceLru.Get(sn)
	if ok {
		p, err := d.products.GetProduct(info.Pk)
		if err != nil {
			return nil, nil, err
		}
		return info, p, nil
	}
	info, err := d.loadShadowFromDB(sn)
	if err != nil {
		return nil, nil, err
	}
	d.deviceLru.Set(sn, info, info.ShadowVersion)
	p, err := d.products.GetProduct(info.Pk)
	if err != nil {
		return nil, nil, err
	}

	return info, p, err
}

func (d *Devices) loadShadowFromDB(sn string) (*Device, error) {
	//数据库加载
	s, err := model.ShadowGetBySn(nil, sn)
	if err != nil {
		return nil, err
	}

	sInfo := &define.Shadow{}
	err = json.Unmarshal([]byte(s.Shadow), sInfo)
	if err != nil {
		return nil, err
	}

	info := &Device{
		PSn:           s.PSn,
		Sn:            sn,
		Group:         utils.GetGroupId(sn),
		Pk:            s.Pk,
		Shadow:        sInfo,
		ShadowVersion: s.LastVersion,
	}
	return info, nil
}
