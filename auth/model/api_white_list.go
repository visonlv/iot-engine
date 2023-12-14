package model

import (
	"time"

	"github.com/visonlv/go-vkit/mysqlx"
	"github.com/visonlv/go-vkit/utilsx"
	"gorm.io/gorm"
)

// api_white_list
var (
	api_white_list_model = &ApiWhiteListModel{}
)

type ApiWhiteListModel struct {
	Id         string    `gorm:"primaryKey;type:varchar(64);comment:主键id"`
	CreatedAt  time.Time // 创建时间
	UpdatedAt  time.Time // 更新时间
	CreateUser string    `gorm:"type:varchar(128);comment:创建用户"`
	UpdateUser string    `gorm:"type:varchar(128);comment:更新用户"`
	IsDelete   int       `gorm:"type:tinyint;not null;default:0;comment:删除状态 0正常 1删除"`

	Name string `gorm:"type:varchar(128);comment:接口名称"`
	Path string `gorm:"type:varchar(128);comment:接口路径"`
}

func (*ApiWhiteListModel) TableName() string {
	return "t_api_white_list"
}

func (a *ApiWhiteListModel) BeforeCreate(tx *gorm.DB) error {
	a.Id = utilsx.GenUuid()
	return nil
}
func ApiWhiteListAdd(tx *mysqlx.MysqlClient, m *ApiWhiteListModel) error {
	if err := getTx(tx).Model(api_white_list_model).Insert(m); err != nil {
		return err
	}
	return nil
}

func ApiWhiteListGet(tx *mysqlx.MysqlClient, id string) (*ApiWhiteListModel, error) {
	item := &ApiWhiteListModel{}
	result := getTx(tx).Where("id = ? AND is_delete = ?", id, 0).First(item)
	return item, result.GetDB().Error
}
func ApiWhiteListDel(tx *mysqlx.MysqlClient, id string) error {
	result := getTx(tx).Model(api_white_list_model).Where("id = ?", id).Update("is_delete", 1)
	return result.GetDB().Error
}
func ApiWhiteListUpdate(tx *mysqlx.MysqlClient, m *ApiWhiteListModel) (*ApiWhiteListModel, error) {
	err := getTx(tx).UpdateEx(m)
	return m, err
}
func ApiWhiteListPage(tx *mysqlx.MysqlClient, pageIndex int32, pageSize int32, name string) (list []*ApiWhiteListModel, total int32, err error) {
	query := getTx(tx).Model(api_white_list_model).Where("is_delete = 0")
	if name != "" {
		query = query.Where("name like ?", "%"+name+"%")
	}
	err = query.Order("created_at desc").FindPage(pageIndex, pageSize, &list, &total)
	return
}
func ApiWhiteListGetByName(tx *mysqlx.MysqlClient, name string) (*ApiWhiteListModel, bool, error) {
	item := &ApiWhiteListModel{}
	exist, err := getTx(tx).Where("name = ? AND is_delete = ?", name, 0).FindFirst(item)
	return item, exist, err
}
func ApiWhiteListGetByPath(tx *mysqlx.MysqlClient, path string) (*ApiWhiteListModel, bool, error) {
	item := &ApiWhiteListModel{}
	exist, err := getTx(tx).Where("path = ? AND is_delete = ?", path, 0).FindFirst(item)
	return item, exist, err
}
