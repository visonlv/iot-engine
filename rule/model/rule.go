package model

import (
	"time"

	"github.com/visonlv/go-vkit/mysqlx"
	"gorm.io/gorm"
)

// 规则表
var (
	rule_model = &RuleModel{}
)

type RuleModel struct {
	Id         string    `gorm:"primaryKey;type:varchar(64);comment:主键id"`
	CreatedAt  time.Time `gorm:"type:TIMESTAMP;comment:创建时间"` // 创建时间
	UpdatedAt  time.Time `gorm:"type:TIMESTAMP;comment:更新时间"` // 更新时间
	CreateUser string    `gorm:"type:varchar(128);comment:创建用户"`
	UpdateUser string    `gorm:"type:varchar(128);comment:更新用户"`
	IsDelete   int       `gorm:"type:tinyint;not null;default:0;comment:删除状态 0正常 1删除"`

	Name        string `gorm:"type:varchar(64);comment:场景名称"`
	TriggerType string `gorm:"type:varchar(64);comment:触发器类型"`
	Trigger     string `gorm:"type:text;comment:触发器数据"`
	Action      string `gorm:"type:text;comment:动作序列"`
	Desc        string `gorm:"type:varchar(64);comment:描述"`
}

func (a *RuleModel) BeforeCreate(tx *gorm.DB) error {
	return nil
}

func (*RuleModel) TableName() string {
	return "t_rule"
}

func RuleAdd(tx *mysqlx.MysqlClient, m *RuleModel) error {
	if err := getTx(tx).Model(rule_model).Insert(m); err != nil {
		return err
	}
	return nil
}

func RuleGet(tx *mysqlx.MysqlClient, id string) (*RuleModel, error) {
	item := &RuleModel{}
	result := getTx(tx).Where("id = ? AND is_delete = ?", id, 0).First(item)
	return item, result.GetDB().Error
}

func RuleGetBySn(tx *mysqlx.MysqlClient, sn string) (*RuleModel, error) {
	item := &RuleModel{}
	result := getTx(tx).Where("sn = ? AND is_delete = ?", sn, 0).First(item)
	return item, result.GetDB().Error
}

func RuleDel(tx *mysqlx.MysqlClient, id string) error {
	result := getTx(tx).Model(rule_model).Where("id = ?", id).Update("is_delete", 1)
	return result.GetDB().Error
}

func RuleUpdate(tx *mysqlx.MysqlClient, m *RuleModel) error {
	return getTx(tx).UpdateEx(m)
}

func RulePage(tx *mysqlx.MysqlClient, pageIndex int32, pageSize int32, pk, name, sn, pId string) ([]*RuleModel, int32, error) {
	items := []*RuleModel{}
	var total int32
	query := getTx(tx).Model(rule_model).Where("is_delete = ?", 0)
	if name != "" {
		query = query.Where("name like ?", "%"+name+"%")
	}
	if pk != "" {
		query = query.Where("pk = ?", pk)
	}
	if sn != "" {
		query = query.Where("sn = ?", sn)
	}
	if pId != "" {
		query = query.Where("p_id = ?", pId)
	}
	err := query.Order("created_at desc").FindPage(pageIndex, pageSize, &items, &total)
	return items, total, err
}

func RuleList(tx *mysqlx.MysqlClient, pk, name, sn, pId string) (list []*RuleModel, err error) {
	query := getTx(tx).Model(rule_model).Where("is_delete = ?", 0)
	if name != "" {
		query = query.Where("name like ?", "%"+name+"%")
	}
	if pk != "" {
		query = query.Where("pk = ?", pk)
	}
	if sn != "" {
		query = query.Where("sn = ?", sn)
	}
	if pId != "" {
		query = query.Where("p_id = ?", pId)
	}
	err = query.FindList(&list)
	return
}
