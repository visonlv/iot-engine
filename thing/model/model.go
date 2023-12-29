package model

import (
	"github.com/visonlv/go-vkit/mysqlx"
	"github.com/visonlv/iot-engine/thing/app"
)

// InitTable 初始化数据库表
func InitTable() {
	// 自动建表
	app.Mysql.GetDB().AutoMigrate(&DeviceModel{}, &ProductModel{}, &ProductModelModel{}, &RuleModel{}, &NotifyLogModel{}, &NotifyTemplateModel{}, &NotifyConfigModel{})
}

func getTx(tx *mysqlx.MysqlClient) *mysqlx.MysqlClient {
	if tx == nil {
		return app.Mysql
	}
	return tx
}
