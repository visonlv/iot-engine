package model

import (
	"time"

	"github.com/visonlv/go-vkit/mysqlx"
	"github.com/visonlv/go-vkit/utilsx"
	"gorm.io/gorm"
)

// 应用
var (
	app_model = &AppModel{}
)

type AppModel struct {
	Id         string    `gorm:"primaryKey;type:varchar(64);comment:主键id"`
	CreatedAt  time.Time // 创建时间
	UpdatedAt  time.Time // 更新时间
	CreateUser string    `gorm:"type:varchar(128);comment:创建用户"`
	UpdateUser string    `gorm:"type:varchar(128);comment:更新用户"`
	IsDelete   int       `gorm:"type:tinyint;not null;default:0;comment:删除状态 0正常 1删除"`

	Code     string `gorm:"type:varchar(64);comment:应用代码(唯一)"`
	Name     string `gorm:"type:varchar(64);comment:应用名称"`
	Describe string `gorm:"type:varchar(64);comment:应用描述"`
}

func (*AppModel) TableName() string {
	return "t_app"
}

func (a *AppModel) BeforeCreate(tx *gorm.DB) error {
	a.Id = utilsx.GenUuid()
	return nil
}

// 添加应用
func AppAdd(tx *mysqlx.MysqlClient, m *AppModel) error {
	if err := getTx(tx).Model(app_model).Insert(m); err != nil {
		return err
	}
	return nil
}
func AppGet(tx *mysqlx.MysqlClient, id string) (*AppModel, error) {
	item := &AppModel{}
	result := getTx(tx).Where("id = ? AND is_delete = ?", id, 0).First(item)
	return item, result.GetDB().Error
}
func AppDel(tx *mysqlx.MysqlClient, id string) error {
	result := getTx(tx).Model(app_model).Where("id = ?", id).Update("is_delete", 1)
	if result.GetDB().Error == nil {
		return nil
	}
	return result.GetDB().Error
}
func AppUpdate(tx *mysqlx.MysqlClient, app *AppModel) error {
	return getTx(tx).Model(app_model).UpdateEx(app)
}
func AppPage(tx *mysqlx.MysqlClient, pageIndex int32, pageSize int32, name string) ([]*AppModel, int32, error) {
	items := []*AppModel{}
	var total int32
	query := getTx(tx).Model(app_model).Where("is_delete = ? ", 0)
	if name != "" {
		query = query.Where("name like ?", "%"+name+"%")
	}
	err := query.Order("created_at desc").FindPage(pageIndex, pageSize, &items, &total)
	return items, total, err
}
func AppGetByCode(tx *mysqlx.MysqlClient, code string) (*AppModel, bool, error) {
	item := &AppModel{}
	exist, err := getTx(tx).Where("code = ? AND is_delete = ?", code, 0).FindFirst(item)
	return item, exist, err
}
