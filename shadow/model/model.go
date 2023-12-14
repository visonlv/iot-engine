package model

import (
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/visonlv/go-vkit/mysqlx"
	"github.com/visonlv/iot-engine/shadow/app"
)

// InitTable 初始化数据库表
func InitTable() {
	// 自动建表
	app.Mysql.GetDB().AutoMigrate(&ShadowModel{})
}

func getTx(tx *mysqlx.MysqlClient) *mysqlx.MysqlClient {
	if tx == nil {
		return app.Mysql
	}
	return tx
}

type OrderBy struct {
	Filed string `json:"filed" form:"filed"` //要排序的字段名
	Sort  int64  `json:"sort" form:"sort"`   //排序的方式：0 OrderAes、1 OrderDesc
}

type PageInfo struct {
	TimeStart int64     `json:"timeStart"`
	TimeEnd   int64     `json:"timeEnd"`
	Page      int64     `json:"page" form:"page"`       // 页码
	Size      int64     `json:"size" form:"size"`       // 每页大小
	Orders    []OrderBy `json:"orderBy" form:"orderBy"` // 排序信息
}

func (p PageInfo) GetTimeStart() time.Time {
	return time.UnixMilli(p.TimeStart)
}
func (p PageInfo) GetTimeEnd() time.Time {
	return time.UnixMilli(p.TimeEnd)
}

func (p PageInfo) FmtWhere(sql sq.SelectBuilder) sq.SelectBuilder {
	if p.TimeStart != 0 {
		sql = sql.Where(sq.GtOrEq{"ts": p.GetTimeStart()})
	}
	if p.TimeEnd != 0 {
		sql = sql.Where(sq.LtOrEq{"ts": p.GetTimeEnd()})
	}
	return sql
}

func (p PageInfo) FmtSql(sql sq.SelectBuilder) sq.SelectBuilder {
	if p.TimeStart != 0 {
		sql = sql.Where("ts>=?", p.GetTimeStart())
	}
	if p.TimeEnd != 0 {
		sql = sql.Where("ts<=?", p.GetTimeEnd())
	}
	if p.Size != 0 {
		sql = sql.Limit(uint64(p.GetLimit()))
		if p.Page != 0 {
			sql = sql.Offset(uint64(p.GetOffset()))
		}
	}
	return sql
}

func (p PageInfo) GetLimit() int64 {
	return p.Size
}

func (p PageInfo) GetOffset() int64 {
	if p.Page == 0 {
		return 0
	}
	return p.Size * (p.Page - 1)
}
