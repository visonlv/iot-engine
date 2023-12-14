package app

import (
	"github.com/visonlv/iot-engine/proxy/config"
)

var (
	Cfg *config.Config
)

func Init(cfgPath string) {
	initConfig(cfgPath)
}

// 初始化配置， 第一步运行
func initConfig(fpath string) {
	if cfg, err := config.LoadConfig(fpath); err == nil {
		Cfg = cfg
	} else {
		panic(err)
	}
}
