package model

import (
	"time"

	"github.com/visonlv/go-vkit/mysqlx"
	"github.com/visonlv/go-vkit/utilsx"
	"gorm.io/gorm"
)

// 产品表
var (
	product_model = &ProductModel{}
)

type ProductModel struct {
	Id         string    `gorm:"primaryKey;type:varchar(64);comment:主键id"`
	CreatedAt  time.Time `gorm:"type:TIMESTAMP;comment:创建时间"` // 创建时间
	UpdatedAt  time.Time `gorm:"type:TIMESTAMP;comment:更新时间"` // 更新时间
	CreateUser string    `gorm:"type:varchar(128);comment:创建用户"`
	UpdateUser string    `gorm:"type:varchar(128);comment:更新用户"`
	IsDelete   int       `gorm:"type:tinyint;not null;default:0;comment:删除状态 0正常 1删除"`

	Name      string `gorm:"type:varchar(64);comment:产品名称"`
	Model     string `gorm:"type:varchar(64);comment:产品型号"`
	Pk        string `gorm:"type:varchar(64);comment:产品KEY"`
	Transform string `gorm:"type:varchar(64);comment:传输类型 model 物模型 raw 透传"`
	Protocol  string `gorm:"type:varchar(64);comment:协议 mqtt3 mqtt5 websocket coap"`
	Type      string `gorm:"type:varchar(64);comment:产品类型 direct 直连 gateway 网关 child 子设备"`
	Desc      string `gorm:"type:varchar(64);comment:描述"`
}

func (a *ProductModel) BeforeCreate(tx *gorm.DB) error {
	a.Id = utilsx.GenUuid()
	return nil
}

func (*ProductModel) TableName() string {
	return "t_product"
}

func ProductAdd(tx *mysqlx.MysqlClient, m *ProductModel) error {
	if err := getTx(tx).Model(product_model).Insert(m); err != nil {
		return err
	}
	return nil
}

func ProductGet(tx *mysqlx.MysqlClient, id string) (*ProductModel, error) {
	item := &ProductModel{}
	result := getTx(tx).Where("id = ? AND is_delete = ?", id, 0).First(item)
	return item, result.GetDB().Error
}

func ProductGetByPk(tx *mysqlx.MysqlClient, pk string) (*ProductModel, error) {
	item := &ProductModel{}
	result := getTx(tx).Where("pk = ? AND is_delete = ?", pk, 0).First(item)
	return item, result.GetDB().Error
}

func ProductGetByModel(tx *mysqlx.MysqlClient, model string) (*ProductModel, error) {
	item := &ProductModel{}
	result := getTx(tx).Where("model = ? AND is_delete = ?", model, 0).First(item)
	return item, result.GetDB().Error
}

func ProductDel(tx *mysqlx.MysqlClient, id string) error {
	result := getTx(tx).Model(product_model).Where("id = ?", id).Update("is_delete", 1)
	return result.GetDB().Error
}

func ProductUpdate(tx *mysqlx.MysqlClient, m *ProductModel) error {
	return getTx(tx).UpdateEx(m)
}

func ProductPage(tx *mysqlx.MysqlClient, pageIndex int32, pageSize int32, name, model, pk, ptype string) ([]*ProductModel, int32, error) {
	items := []*ProductModel{}
	var total int32
	query := getTx(tx).Model(product_model).Where("is_delete = ?", 0)
	if name != "" {
		query = query.Where("name like ?", "%"+name+"%")
	}
	if model != "" {
		query = query.Where("model = ?", model)
	}
	if pk != "" {
		query = query.Where("pk = ?", pk)
	}
	if ptype != "" {
		query = query.Where("type = ?", ptype)
	}
	err := query.Order("created_at desc").FindPage(pageIndex, pageSize, &items, &total)
	return items, total, err
}

func ProductList(tx *mysqlx.MysqlClient, name, model, pk, ptype string) (list []*ProductModel, err error) {
	query := getTx(tx).Model(product_model).Where("is_delete = ?", 0)
	if name != "" {
		query = query.Where("name like ?", "%"+name+"%")
	}
	if model != "" {
		query = query.Where("model = ?", model)
	}
	if pk != "" {
		query = query.Where("pk = ?", pk)
	}
	if ptype != "" {
		query = query.Where("type = ?", ptype)
	}
	err = query.FindList(&list)
	return
}

func ProductExistByPk(tx *mysqlx.MysqlClient, pk string) (bool, error) {
	var c int64 = 0
	result := getTx(tx).Model(product_model).Where("pk = ? AND is_delete = ?", pk, 0).Count(&c)
	return c > 0, result.GetDB().Error
}

func ProductGetInPks(tx *mysqlx.MysqlClient, pks []string) (list []*ProductModel, err error) {
	err = getTx(tx).Model(product_model).Where("pk in ? and is_delete = ?", pks, 0).FindList(&list)
	return
}
