package model

import (
	"time"

	"github.com/visonlv/go-vkit/mysqlx"
	"github.com/visonlv/go-vkit/utilsx"
	"gorm.io/gorm"
)

// 产品物模型定义表
var (
	product_model_model = &ProductModelModel{}
)

type ProductModelModel struct {
	Id         string    `gorm:"primaryKey;type:varchar(64);comment:主键id"`
	CreatedAt  time.Time `gorm:"type:TIMESTAMP;comment:创建时间"` // 创建时间
	UpdatedAt  time.Time `gorm:"type:TIMESTAMP;comment:更新时间"` // 更新时间
	CreateUser string    `gorm:"type:varchar(128);comment:创建用户"`
	UpdateUser string    `gorm:"type:varchar(128);comment:更新用户"`
	IsDelete   int       `gorm:"type:tinyint;not null;default:0;comment:删除状态 0正常 1删除"`

	ProductId string `gorm:"type:varchar(64);comment:产品Id"`
	Name      string `gorm:"type:varchar(64);comment:名称"`
	Code      string `gorm:"type:varchar(64);comment:标识"`
	Type      string `gorm:"type:varchar(64);comment:定义类型 property event service"`
	Desc      string `gorm:"type:varchar(64);comment:描述"`
	ModelDef  string `gorm:"type:text;comment:物模型详细"`
	IsSys     int    `gorm:"type:tinyint;not null;default:0;comment:是否系统模型 1 是 0 否"`
}

func (a *ProductModelModel) BeforeCreate(tx *gorm.DB) error {
	a.Id = utilsx.GenUuid()
	return nil
}

func (*ProductModelModel) TableName() string {
	return "t_product_model"
}

func ProductModelAdd(tx *mysqlx.MysqlClient, m *ProductModelModel) error {
	if err := getTx(tx).Model(product_model_model).Insert(m); err != nil {
		return err
	}
	return nil
}

func ProductModelAddBatch(tx *mysqlx.MysqlClient, m []*ProductModelModel) error {
	if err := getTx(tx).Model(product_model_model).Insert(m); err != nil {
		return err
	}
	return nil
}

func ProductModelGet(tx *mysqlx.MysqlClient, id string) (*ProductModelModel, error) {
	item := &ProductModelModel{}
	result := getTx(tx).Where("id = ? AND is_delete = ?", id, 0).First(item)
	return item, result.GetDB().Error
}

func ProductModelGetByProductIdAndCodeAndType(tx *mysqlx.MysqlClient, productId, code, ptype string) (*ProductModelModel, error) {
	item := &ProductModelModel{}
	result := getTx(tx).Where("product_id = ? AND code = ? AND type = ? AND is_delete = ?", productId, code, ptype, 0).First(item)
	return item, result.GetDB().Error
}

func ProductModelDel(tx *mysqlx.MysqlClient, id string) error {
	result := getTx(tx).Model(product_model_model).Where("id = ?", id).Update("is_delete", 1)
	return result.GetDB().Error
}

func ProductModelDelByProductId(tx *mysqlx.MysqlClient, productId string) error {
	result := getTx(tx).Model(product_model_model).Where("product_id = ?", productId).Update("is_delete", 1)
	return result.GetDB().Error
}

func ProductModelUpdateByProductId(tx *mysqlx.MysqlClient, productId string, list []*ProductModelModel) error {
	err := getTx(tx).Transaction(func(tx *mysqlx.MysqlClient) error {
		if err := ProductModelDelByProductId(tx, productId); err != nil {
			return err
		}
		if err := ProductModelAddBatch(tx, list); err != nil {
			return err
		}
		return nil
	})
	return err
}

func ProductModelUpdate(tx *mysqlx.MysqlClient, m *ProductModelModel) error {
	return getTx(tx).UpdateEx(m)
}

func ProductModelPage(tx *mysqlx.MysqlClient, pageIndex int32, pageSize int32, productId, name, code, ptype string) ([]*ProductModelModel, int32, error) {
	items := []*ProductModelModel{}
	var total int32
	query := getTx(tx).Model(product_model_model).Where("is_delete = ?", 0)
	if productId != "" {
		query = query.Where("product_id = ?", productId)
	}
	if name != "" {
		query = query.Where("name like ?", "%"+name+"%")
	}
	if code != "" {
		query = query.Where("code = ?", code)
	}
	if ptype != "" {
		query = query.Where("type = ?", ptype)
	}
	err := query.Order("created_at desc").FindPage(pageIndex, pageSize, &items, &total)
	return items, total, err
}

func ProductModelList(tx *mysqlx.MysqlClient, productId, name, code, ptype string) (list []*ProductModelModel, err error) {
	query := getTx(tx).Model(product_model_model).Where("is_delete = ?", 0)
	if productId != "" {
		query = query.Where("product_id = ?", productId)
	}
	if name != "" {
		query = query.Where("name like ?", "%"+name+"%")
	}
	if code != "" {
		query = query.Where("code = ?", code)
	}
	if ptype != "" {
		query = query.Where("type = ?", ptype)
	}
	err = query.FindList(&list)
	return
}
