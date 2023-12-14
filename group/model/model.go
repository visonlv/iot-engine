package model

import (
	"github.com/visonlv/go-vkit/mysqlx"
	"github.com/visonlv/iot-engine/group/app"
)

// InitTable 初始化数据库表
func InitTable() {
	// 自动建表
	app.Mysql.GetDB().AutoMigrate(&CategoryModel{})
}

func getTx(tx *mysqlx.MysqlClient) *mysqlx.MysqlClient {
	if tx == nil {
		return app.Mysql
	}
	return tx
}
