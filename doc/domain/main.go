package main

import (
	"fmt"
)

// 示例
func main() {
	// 定义一个物模型
	model := Model{
		Properties: map[string]PropertyDef{
			"speed": {
				Name:        "行驶速度",
				Description: "",
				Unit:        "km/h",
				Type:        FLOAT,
				Writable:    true,
				Min:         0,
				Max:         120,
			},
		},
		Services: map[string]ServiceDef{
			"selfClean": {
				Name:        "自清洗",
				Description: "",
				Parameters: []ServiceParameterDef{
					{
						Name:        "清洗力度",
						Description: "",
						Type:        INT,
						Options:     []interface{}{1, 2, 3},
					},
				},
			},
		},
		Events: map[string]EventDef{
			"selfCleanFinished": {
				Name:        "自清洗完成",
				Description: "",
				Parameters: []EventParameterDef{
					{
						Name:        "cost",
						Description: "清洗耗时",
						Type:        INT,
					},
				},
			},
		},
	}
	// 创建一个产品
	product := Product{
		Id:          "1",
		Name:        "站岗",
		Description: "",
		Protocol:    MQTT3,
		Type:        DIRECT,
		Model:       model,
	}
	// 创建一个设备
	device := Device{
		Id:      "1",
		Name:    "站岗1号",
		Sn:      "ZG1",
		Key:     "私钥",
		Product: product,
		Shadow:  Shadow{},
	}

	fmt.Println(device)
}
