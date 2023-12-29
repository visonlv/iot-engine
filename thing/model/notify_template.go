package model

import (
	"time"

	"github.com/visonlv/go-vkit/mysqlx"
	"github.com/visonlv/go-vkit/utilsx"
	"gorm.io/gorm"
)

// 通知模板表
var (
	notify_template_model = &NotifyTemplateModel{}
)

type NotifyTemplateModel struct {
	Id         string    `gorm:"primaryKey;type:varchar(64);comment:主键id"`
	CreatedAt  time.Time `gorm:"type:TIMESTAMP;comment:创建时间"` // 创建时间
	UpdatedAt  time.Time `gorm:"type:TIMESTAMP;comment:更新时间"` // 更新时间
	CreateUser string    `gorm:"type:varchar(128);comment:创建用户"`
	UpdateUser string    `gorm:"type:varchar(128);comment:更新用户"`
	IsDelete   int       `gorm:"type:tinyint;not null;default:0;comment:删除状态 0正常 1删除"`

	Name           string `gorm:"type:varchar(64);comment:名称"`
	NotifyConfigId string `gorm:"type:varchar(64);comment:通知配置id"`
	NotifyType     string `gorm:"type:varchar(64);comment:通知类型"`
	NotifyTemplate string `gorm:"type:text;comment:通知模板"`
	Desc           string `gorm:"type:varchar(64);comment:描述"`
}

func (a *NotifyTemplateModel) BeforeCreate(tx *gorm.DB) error {
	a.Id = utilsx.GenUuid()
	return nil
}

func (*NotifyTemplateModel) TableName() string {
	return "t_notify_template"
}

func NotifyTemplateAdd(tx *mysqlx.MysqlClient, m *NotifyTemplateModel) error {
	if err := getTx(tx).Model(notify_template_model).Insert(m); err != nil {
		return err
	}
	return nil
}

func NotifyTemplateGet(tx *mysqlx.MysqlClient, id string) (*NotifyTemplateModel, error) {
	item := &NotifyTemplateModel{}
	result := getTx(tx).Where("id = ? AND is_delete = ?", id, 0).First(item)
	return item, result.GetDB().Error
}

func NotifyTemplateByName(tx *mysqlx.MysqlClient, name string) (*NotifyTemplateModel, error) {
	item := &NotifyTemplateModel{}
	result := getTx(tx).Where("name = ? AND is_delete = ?", name, 0).First(item)
	return item, result.GetDB().Error
}

func NotifyTemplateDel(tx *mysqlx.MysqlClient, id string) error {
	result := getTx(tx).Model(notify_template_model).Where("id = ?", id).Update("is_delete", 1)
	return result.GetDB().Error
}

func NotifyTemplateUpdate(tx *mysqlx.MysqlClient, m *NotifyTemplateModel) error {
	return getTx(tx).UpdateEx(m)
}

func NotifyTemplatePage(tx *mysqlx.MysqlClient, pageIndex int32, pageSize int32, name, notifyType, notifyConfigId string) ([]*NotifyTemplateModel, int32, error) {
	items := []*NotifyTemplateModel{}
	var total int32
	query := getTx(tx).Model(notify_template_model).Where("is_delete = ?", 0)
	if name != "" {
		query = query.Where("name like ?", "%"+name+"%")
	}
	if notifyType != "" {
		query = query.Where("notify_type = ?", notifyType)
	}
	if notifyConfigId != "" {
		query = query.Where("notify_config_id = ?", notifyConfigId)
	}

	err := query.Order("created_at desc").FindPage(pageIndex, pageSize, &items, &total)
	return items, total, err
}

func NotifyTemplateList(tx *mysqlx.MysqlClient, name, notifyType, notifyConfigId string) (list []*NotifyTemplateModel, err error) {
	query := getTx(tx).Model(notify_template_model).Where("is_delete = ?", 0)
	if name != "" {
		query = query.Where("name like ?", "%"+name+"%")
	}
	if notifyType != "" {
		query = query.Where("notify_type = ?", notifyType)
	}
	if notifyConfigId != "" {
		query = query.Where("notify_config_id = ?", notifyConfigId)
	}
	err = query.FindList(&list)
	return
}

func NotifyTemplateListInIds(tx *mysqlx.MysqlClient, ids []string) (list []*NotifyTemplateModel, err error) {
	err = getTx(tx).Model(notify_template_model).Where("id in ? and is_delete = ?", ids, 0).FindList(&list)
	return
}

func NotifyTemplateCountByConfigId(tx *mysqlx.MysqlClient, notifyConfigId string) (count int64, err error) {
	result := getTx(tx).Model(notify_template_model).Where("notify_config_id = ? AND is_delete = ?", notifyConfigId, 0).Count(&count)
	err = result.GetDB().Error
	return
}
