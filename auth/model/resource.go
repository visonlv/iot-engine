package model

import (
	"time"

	"github.com/visonlv/go-vkit/mysqlx"
	"github.com/visonlv/go-vkit/utilsx"
	"gorm.io/gorm"
)

// resource
var (
	resource_model = &ResourceModel{}
)

type ResourceModel struct {
	Id         string    `gorm:"primaryKey;type:varchar(64);comment:主键id"`
	CreatedAt  time.Time // 创建时间
	UpdatedAt  time.Time // 更新时间
	CreateUser string    `gorm:"type:varchar(128);comment:创建用户"`
	UpdateUser string    `gorm:"type:varchar(128);comment:更新用户"`
	IsDelete   int       `gorm:"type:tinyint;not null;default:0;comment:删除状态 0正常 1删除"`

	Name     string `gorm:"type:varchar(128);comment:资源名称"`
	ParentId string `gorm:"type:varchar(128);comment:父资源"`
	Type     string `gorm:"type:varchar(128);comment:资源类型(api/menu/rule)"`
	Content  string `gorm:"type:text;comment:资源内容"`
	Property string `gorm:"type:varchar(128);comment:资源属性"`
}

func (*ResourceModel) TableName() string {
	return "t_resource"
}

func (a *ResourceModel) BeforeCreate(tx *gorm.DB) error {
	a.Id = utilsx.GenUuid()
	return nil
}
func ResourceAdd(tx *mysqlx.MysqlClient, m *ResourceModel) error {
	if err := getTx(tx).Model(resource_model).Insert(m); err != nil {
		return err
	}
	return nil
}

func ResourceGet(tx *mysqlx.MysqlClient, id string) (*ResourceModel, error) {
	item := &ResourceModel{}
	result := getTx(tx).Where("id = ? AND is_delete = ?", id, 0).First(item)
	return item, result.GetDB().Error
}
func ResourceDel(tx *mysqlx.MysqlClient, id string) error {
	result := getTx(tx).Model(resource_model).Where("id = ?", id).Update("is_delete", 1)
	return result.GetDB().Error
}
func ResourceUpdate(tx *mysqlx.MysqlClient, m *ResourceModel) (*ResourceModel, error) {
	err := getTx(tx).UpdateEx(m)
	return m, err
}
func ResourcePage(tx *mysqlx.MysqlClient, pageIndex int32, pageSize int32, name, resType, content string) (list []*ResourceModel, total int32, err error) {
	query := getTx(tx).Model(resource_model).Where("is_delete = 0")
	if name != "" {
		query = query.Where("name like ?", "%"+name+"%")
	}
	if resType != "" {
		query = query.Where("type = ?", resType)
	}
	if content != "" {
		query = query.Where("content like ?", "%"+content+"%")
	}
	query = query.Order("created_at desc ,id")
	err = query.FindPage(pageIndex, pageSize, &list, &total)
	return
}

func ResourceListByIds(tx *mysqlx.MysqlClient, ids []string) (list []*ResourceModel, err error) {
	items := []*ResourceModel{}
	err = getTx(tx).Where("id in (?) and is_delete = ?", ids, 0).FindList(&items)
	return items, err
}
func ResourceGetByNameAndType(tx *mysqlx.MysqlClient, name, resType string) (*ResourceModel, bool, error) {
	item := &ResourceModel{}
	exist, err := getTx(tx).Where("name = ? AND type = ? AND is_delete = ?", name, resType, 0).FindFirst(item)
	return item, exist, err
}
