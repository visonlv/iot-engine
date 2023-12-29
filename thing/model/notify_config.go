package model

import (
	"time"

	"github.com/visonlv/go-vkit/mysqlx"
	"github.com/visonlv/go-vkit/utilsx"
	"gorm.io/gorm"
)

// 通知配置表
var (
	notify_config_model = &NotifyConfigModel{}
)

type NotifyConfigModel struct {
	Id         string    `gorm:"primaryKey;type:varchar(64);comment:主键id"`
	CreatedAt  time.Time `gorm:"type:TIMESTAMP;comment:创建时间"` // 创建时间
	UpdatedAt  time.Time `gorm:"type:TIMESTAMP;comment:更新时间"` // 更新时间
	CreateUser string    `gorm:"type:varchar(128);comment:创建用户"`
	UpdateUser string    `gorm:"type:varchar(128);comment:更新用户"`
	IsDelete   int       `gorm:"type:tinyint;not null;default:0;comment:删除状态 0正常 1删除"`

	Name         string `gorm:"type:varchar(64);comment:名称"`
	NotifyType   string `gorm:"type:varchar(64);comment:通知类型"`
	NotifyConfig string `gorm:"type:text;comment:通知内容"`
	Desc         string `gorm:"type:varchar(64);comment:描述"`
}

func (a *NotifyConfigModel) BeforeCreate(tx *gorm.DB) error {
	a.Id = utilsx.GenUuid()
	return nil
}

func (*NotifyConfigModel) TableName() string {
	return "t_notify_config"
}

func NotifyConfigAdd(tx *mysqlx.MysqlClient, m *NotifyConfigModel) error {
	if err := getTx(tx).Model(notify_config_model).Insert(m); err != nil {
		return err
	}
	return nil
}

func NotifyConfigGet(tx *mysqlx.MysqlClient, id string) (*NotifyConfigModel, error) {
	item := &NotifyConfigModel{}
	result := getTx(tx).Where("id = ? AND is_delete = ?", id, 0).First(item)
	return item, result.GetDB().Error
}

func NotifyConfigByName(tx *mysqlx.MysqlClient, name string) (*NotifyConfigModel, error) {
	item := &NotifyConfigModel{}
	result := getTx(tx).Where("name = ? AND is_delete = ?", name, 0).First(item)
	return item, result.GetDB().Error
}

func NotifyConfigDel(tx *mysqlx.MysqlClient, id string) error {
	result := getTx(tx).Model(notify_config_model).Where("id = ?", id).Update("is_delete", 1)
	return result.GetDB().Error
}

func NotifyConfigUpdate(tx *mysqlx.MysqlClient, m *NotifyConfigModel) error {
	return getTx(tx).UpdateEx(m)
}

func NotifyConfigPage(tx *mysqlx.MysqlClient, pageIndex int32, pageSize int32, name, notifyType string) ([]*NotifyConfigModel, int32, error) {
	items := []*NotifyConfigModel{}
	var total int32
	query := getTx(tx).Model(notify_config_model).Where("is_delete = ?", 0)
	if name != "" {
		query = query.Where("name like ?", "%"+name+"%")
	}
	if notifyType != "" {
		query = query.Where("notify_type = ?", notifyType)
	}
	err := query.Order("created_at desc").FindPage(pageIndex, pageSize, &items, &total)
	return items, total, err
}

func NotifyConfigList(tx *mysqlx.MysqlClient, name, notifyType string) (list []*NotifyConfigModel, err error) {
	query := getTx(tx).Model(notify_config_model).Where("is_delete = ?", 0)
	if name != "" {
		query = query.Where("name like ?", "%"+name+"%")
	}
	if notifyType != "" {
		query = query.Where("notify_type = ?", notifyType)
	}
	err = query.FindList(&list)
	return
}

func NotifyConfigListInIds(tx *mysqlx.MysqlClient, ids []string) (list []*NotifyConfigModel, err error) {
	err = getTx(tx).Model(notify_config_model).Where("id in ? and is_delete = ?", ids, 0).FindList(&list)
	return
}
