package model

import (
	"time"

	"github.com/visonlv/go-vkit/mysqlx"
	"github.com/visonlv/go-vkit/utilsx"
	"gorm.io/gorm"
)

// 通知日志表
var (
	notify_log_model = &NotifyLogModel{}
)

type NotifyLogModel struct {
	Id         string    `gorm:"primaryKey;type:varchar(64);comment:主键id"`
	CreatedAt  time.Time `gorm:"type:TIMESTAMP;comment:创建时间"` // 创建时间
	UpdatedAt  time.Time `gorm:"type:TIMESTAMP;comment:更新时间"` // 更新时间
	CreateUser string    `gorm:"type:varchar(128);comment:创建用户"`
	UpdateUser string    `gorm:"type:varchar(128);comment:更新用户"`
	IsDelete   int       `gorm:"type:tinyint;not null;default:0;comment:删除状态 0正常 1删除"`

	NotifyType       string `gorm:"type:varchar(64);comment:通知类型"`
	NotifyConfigId   string `gorm:"type:varchar(64);comment:通知配置id"`
	NotifyTemplateId string `gorm:"type:varchar(64);comment:通知模板id"`
	Content          string `gorm:"type:text;comment:通知内容"`
	ResultStatus     int    `gorm:"type:tinyint;comment:执行结果状态"`
	Result           string `gorm:"type:text;comment:执行结果"`
}

func (a *NotifyLogModel) BeforeCreate(tx *gorm.DB) error {
	a.Id = utilsx.GenUuid()
	return nil
}

func (*NotifyLogModel) TableName() string {
	return "t_notify_log"
}

func NotifyLogAdd(tx *mysqlx.MysqlClient, m *NotifyLogModel) error {
	if err := getTx(tx).Model(notify_log_model).Insert(m); err != nil {
		return err
	}
	return nil
}

func NotifyLogGet(tx *mysqlx.MysqlClient, id string) (*NotifyLogModel, error) {
	item := &NotifyLogModel{}
	result := getTx(tx).Where("id = ? AND is_delete = ?", id, 0).First(item)
	return item, result.GetDB().Error
}

func NotifyLogByName(tx *mysqlx.MysqlClient, name string) (*NotifyLogModel, error) {
	item := &NotifyLogModel{}
	result := getTx(tx).Where("name = ? AND is_delete = ?", name, 0).First(item)
	return item, result.GetDB().Error
}

func NotifyLogDel(tx *mysqlx.MysqlClient, id string) error {
	result := getTx(tx).Model(notify_log_model).Where("id = ?", id).Update("is_delete", 1)
	return result.GetDB().Error
}

func NotifyLogUpdate(tx *mysqlx.MysqlClient, m *NotifyLogModel) error {
	return getTx(tx).UpdateEx(m)
}

func NotifyLogPage(tx *mysqlx.MysqlClient, pageIndex int32, pageSize int32, notifyConfigId, notifyTemplateId, notifyType string) ([]*NotifyLogModel, int32, error) {
	items := []*NotifyLogModel{}
	var total int32
	query := getTx(tx).Model(notify_log_model).Where("is_delete = ?", 0)
	if notifyType != "" {
		query = query.Where("notify_type = ?", notifyType)
	}
	if notifyConfigId != "" {
		query = query.Where("notify_config_id = ?", notifyConfigId)
	}
	if notifyTemplateId != "" {
		query = query.Where("notify_template_id = ?", notifyTemplateId)
	}
	err := query.Order("created_at desc").FindPage(pageIndex, pageSize, &items, &total)
	return items, total, err
}

func NotifyLogList(tx *mysqlx.MysqlClient, notifyConfigId, notifyTemplateId, notifyType string) (list []*NotifyLogModel, err error) {
	query := getTx(tx).Model(notify_log_model).Where("is_delete = ?", 0)
	if notifyType != "" {
		query = query.Where("trigger_type = ?", notifyType)
	}
	if notifyConfigId != "" {
		query = query.Where("notify_type = ?", notifyConfigId)
	}
	if notifyTemplateId != "" {
		query = query.Where("notify_template_id = ?", notifyTemplateId)
	}
	err = query.FindList(&list)
	return
}
