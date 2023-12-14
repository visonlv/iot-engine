package model

import (
	"fmt"
	"time"

	"github.com/visonlv/go-vkit/mysqlx"
	"github.com/visonlv/go-vkit/utilsx"
	"gorm.io/gorm"
)

// 角色
var (
	role_model = &RoleModel{}
)

type RoleModel struct {
	Id         string    `gorm:"primaryKey;type:varchar(64);comment:主键id"`
	CreatedAt  time.Time // 创建时间
	UpdatedAt  time.Time // 更新时间
	CreateUser string    `gorm:"type:varchar(128);comment:创建用户"`
	UpdateUser string    `gorm:"type:varchar(128);comment:更新用户"`
	IsDelete   int       `gorm:"type:tinyint;not null;default:0;comment:删除状态 0正常 1删除"`

	Code     string `gorm:"type:varchar(64);comment:角色代码"`
	Rolename string `gorm:"type:varchar(64);comment:角色名称"`
	// AppCode       string `gorm:"type:varchar(64);comment:应用代码"`
	// IsMultiSignOn int    `gorm:"type:tinyint;comment:是否支持多点登录 0-单点登录 1-多点登录;default:0"`
}

func (*RoleModel) TableName() string {
	return "t_role"
}

func (a *RoleModel) BeforeCreate(tx *gorm.DB) error {
	a.Id = utilsx.GenUuid()
	return nil
}

func RoleAdd(tx *mysqlx.MysqlClient, m *RoleModel) error {
	if err := getTx(tx).Model(role_model).Insert(m); err != nil {
		return err
	}
	return nil
}

func RoleGet(tx *mysqlx.MysqlClient, id string) (*RoleModel, error) {
	item := &RoleModel{}
	result := getTx(tx).Where("id = ? AND is_delete = ?", id, 0).First(item)
	return item, result.GetDB().Error
}

func RoleDel(tx *mysqlx.MysqlClient, id string) error {
	result := getTx(tx).Model(role_model).Where("id = ?", id).Update("is_delete", 1)
	return result.GetDB().Error
}

func RoleListByUserId(tx *mysqlx.MysqlClient, userId string) (list []*RoleModel, err error) {
	sql := fmt.Sprintf("select * from t_role where id in(select role_id from t_user_role where user_id = '%s') and is_delete = %d", userId, 0)
	sql += " order by id"
	err = getTx(tx).Raw(sql).FindList(&list)
	return list, err
}

func RoleUpdate(tx *mysqlx.MysqlClient, m *RoleModel) (*RoleModel, error) {
	err := getTx(tx).UpdateEx(m)
	return m, err
}

func RolePage(tx *mysqlx.MysqlClient, pageIndex int32, pageSize int32, name string) (list []*RoleModel, total int32, err error) {
	query := getTx(tx).Model(role_model).Where("is_delete = ?", 0)
	if len(name) != 0 {
		query = query.Where("rolename like ?", "%"+name+"%")
	}
	query.Order("created_at desc").FindPage(pageIndex, pageSize, &list, &total)
	return
}

func RoleListByCodes(tx *mysqlx.MysqlClient, codes []string) ([]*RoleModel, error) {
	list := make([]*RoleModel, 0)
	err := getTx(tx).Where("code IN ? AND is_delete = ?", codes, 0).FindList(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}
func RoleCodeListByUserId(tx *mysqlx.MysqlClient, userId string) ([]string, error) {
	sql := fmt.Sprintf("select code from t_role where id in(select role_id from t_user_role ur inner join t_role r on r.is_delete = %d and ur.user_id = '%s');", 0, userId)
	list := []string{}
	err := getTx(tx).Raw(sql).FindList(&list)
	return list, err
}

func RoleIsExist(tx *mysqlx.MysqlClient, rolecode string) (bool, error) {
	return getTx(tx).Model(role_model).Where("code = ? AND is_delete = ?", rolecode, 0).Exists(&RoleModel{})
}

func RoleNameIsExist(tx *mysqlx.MysqlClient, name string) (bool, error) {
	return getTx(tx).Model(role_model).Where("rolename = ? AND is_delete = ?", name, 0).Exists(&RoleModel{})
}
func RoleGetByCode(tx *mysqlx.MysqlClient, code string) (*RoleModel, error) {
	item := &RoleModel{}
	result := getTx(tx).Where("code = ? AND is_delete = ?", code, 0).First(item)
	return item, result.GetDB().Error
}
