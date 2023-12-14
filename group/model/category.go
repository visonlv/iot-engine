package model

import (
	"time"

	"github.com/visonlv/go-vkit/mysqlx"
	"github.com/visonlv/go-vkit/utilsx"
	"gorm.io/gorm"
)

// 分类表
var (
	category_model = &CategoryModel{}
)

// 结构体
type CategoryModel struct {
	Id         string    `gorm:"primaryKey;type:varchar(64);comment:主键id"`
	CreatedAt  time.Time `gorm:"type:TIMESTAMP;comment:创建时间"` // 创建时间
	UpdatedAt  time.Time `gorm:"type:TIMESTAMP;comment:更新时间"` // 更新时间
	CreateUser string    `gorm:"type:varchar(128);comment:创建用户"`
	UpdateUser string    `gorm:"type:varchar(128);comment:更新用户"`
	IsDelete   int       `gorm:"type:tinyint;not null;default:0;comment:删除状态 0正常 1删除"`

	Name    string `gorm:"type:varchar(64);comment:分类名称;"`
	Code    string `gorm:"type:varchar(64);comment:分类代码;"`
	Content string `gorm:"type:text;comment:内容信息;"`
}

func (a *CategoryModel) BeforeCreate(tx *gorm.DB) error {
	a.Id = utilsx.GenUuid()
	return nil
}

func (*CategoryModel) TableName() string {
	return "t_category"
}

func CategoryAdd(tx *mysqlx.MysqlClient, m *CategoryModel) error {
	if err := getTx(tx).Model(category_model).Insert(m); err != nil {
		return err
	}
	return nil
}

func CategoryGet(tx *mysqlx.MysqlClient, id string) (*CategoryModel, error) {
	item := &CategoryModel{}
	result := getTx(tx).Where("id = ? AND is_delete = ?", id, 0).First(item)
	return item, result.GetDB().Error
}

func CategoryGetByCode(tx *mysqlx.MysqlClient, code string) (*CategoryModel, error) {
	item := &CategoryModel{}
	result := getTx(tx).Where("code = ? AND is_delete = ?", code, 0).First(item)
	return item, result.GetDB().Error
}

func CategoryDel(tx *mysqlx.MysqlClient, id string) error {
	result := getTx(tx).Model(category_model).Where("id = ?", id).Update("is_delete", 1)
	return result.GetDB().Error
}

func CategoryUpdate(tx *mysqlx.MysqlClient, m *CategoryModel) error {
	return getTx(tx).UpdateEx(m)
}

func CategoryPage(tx *mysqlx.MysqlClient, pageIndex int32, pageSize int32) ([]*CategoryModel, int32, error) {
	items := []*CategoryModel{}
	var total int32
	err := getTx(tx).Model(category_model).Where("is_delete = ?", 0).Order("created_at desc").FindPage(pageIndex, pageSize, &items, &total)
	return items, total, err
}

func CategoryList(tx *mysqlx.MysqlClient) (list []*CategoryModel, err error) {
	err = getTx(tx).Model(category_model).Where("is_delete = ?", 0).FindList(&list)
	return
}
