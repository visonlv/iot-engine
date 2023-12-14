package model

import (
	"time"

	"github.com/visonlv/go-vkit/mysqlx"
	"github.com/visonlv/go-vkit/utilsx"
	"gorm.io/gorm"
)

// 登录记录表
var (
	login_record_model = &LoginRecordModel{}
)

// 结构体
type LoginRecordModel struct {
	Id         string    `gorm:"primaryKey;type:varchar(64);comment:主键id"`
	CreatedAt  time.Time // 创建时间
	UpdatedAt  time.Time // 更新时间
	CreateUser string    `gorm:"type:varchar(128);comment:创建用户"`
	UpdateUser string    `gorm:"type:varchar(128);comment:更新用户"`
	IsDelete   int       `gorm:"type:tinyint;not null;default:0;comment:删除状态 0正常 1删除"`

	AppCode  string `gorm:"type:varchar(64);comment:应用代码;"`
	UserId   string `gorm:"type:varchar(64);comment:用户id;"`
	RoleCode string `gorm:"type:varchar(64);comment:角色代码;"`
	Token    string `gorm:"type:text;comment:登录token;"`
}

func (a *LoginRecordModel) BeforeCreate(tx *gorm.DB) error {
	a.Id = utilsx.GenUuid()
	return nil
}

func (*LoginRecordModel) TableName() string {
	return "t_login_record"
}

// 添加应用
func LoginRecordAdd(tx *mysqlx.MysqlClient, m *LoginRecordModel) error {
	if err := getTx(tx).Model(login_record_model).Insert(m); err != nil {
		return err
	}
	return nil
}

func LoginRecordGet(tx *mysqlx.MysqlClient, id string) (*LoginRecordModel, error) {
	item := &LoginRecordModel{}
	result := getTx(tx).Where("id = ? AND is_delete = ?", id, 0).First(item)
	return item, result.GetDB().Error
}

func LoginRecordDel(tx *mysqlx.MysqlClient, id string) error {
	result := getTx(tx).Model(login_record_model).Where("id = ?", id).Update("is_delete", 1)
	if result.GetDB().Error == nil {
		return nil
	}
	return result.GetDB().Error
}

func LoginRecordUpdate(tx *mysqlx.MysqlClient, m *LoginRecordModel) error {
	return getTx(tx).UpdateEx(m)
}

func LoginRecordPage(tx *mysqlx.MysqlClient, pageIndex int32, pageSize int32) ([]*LoginRecordModel, int32, error) {
	items := []*LoginRecordModel{}
	var total int32
	err := getTx(tx).Model(login_record_model).Order("created_at desc").FindPage(pageIndex, pageSize, items, &total)
	return items, total, err
}
func LoginRecordListByUserAndRole(tx *mysqlx.MysqlClient, userId, rolecode string) ([]*LoginRecordModel, error) {
	item := []*LoginRecordModel{}
	err := getTx(tx).Where("user_id = ? AND role_code = ? AND is_delete = ?", userId, rolecode, 0).FindList(&item)
	return item, err
}
func LoginRecordDelByUpdateTime(tx *mysqlx.MysqlClient, expiretime string) {
	getTx(tx).Where("updated_at <= ? ", expiretime).Delete(&LoginRecordModel{})
}
