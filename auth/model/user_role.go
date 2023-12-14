package model

import (
	"time"

	"github.com/visonlv/go-vkit/mysqlx"
	"github.com/visonlv/go-vkit/utilsx"
	"gorm.io/gorm"
)

// 用户角色
var (
	user_role_model = &UserRoleModel{}
)

type UserRoleModel struct {
	Id         string    `gorm:"primaryKey;type:varchar(64);comment:主键id"`
	CreatedAt  time.Time `gorm:"comment:创建时间"` // 在创建时，如果该字段值为F零值，则使用当前时间填充
	UpdatedAt  time.Time `gorm:"comment:更新时间"` // 在创建时该字段值为零值或者在更新时，使用当前时间戳秒数填充
	CreateUser string    `gorm:"type:varchar(128);comment:创建用户"`
	UpdateUser string    `gorm:"type:varchar(128);comment:更新用户"`
	IsDelete   int       `gorm:"type:tinyint;not null;default:0;comment:删除状态 0正常 1删除"`

	UserId string `gorm:"type:varchar(64);comment:用户id"`
	RoleId string `gorm:"type:varchar(64);comment:角色id"`
}

func (*UserRoleModel) TableName() string {
	return "t_user_role"
}

func (a *UserRoleModel) BeforeCreate(tx *gorm.DB) error {
	a.Id = utilsx.GenUuid()
	return nil
}

func UserRoleDelByUserId(tx *mysqlx.MysqlClient, userId string) error {
	result := getTx(tx).Delete(user_role_model, "user_id = ?", userId)
	if result.GetDB().Error != nil {
		return result.GetDB().Error
	}
	return nil
}

func UserRoleAdd(tx *mysqlx.MysqlClient, m *UserRoleModel) error {
	if err := getTx(tx).Model(user_role_model).Insert(m); err != nil {
		return err
	}
	return nil
}

func UserRoleDel(tx *mysqlx.MysqlClient, userId, roleId string) error {
	if err := getTx(tx).Model(user_role_model).Where("user_id=? and role_id = ?", userId, roleId).Delete(user_role_model).GetDB().Error; err != nil {
		return err
	}
	return nil
}

func UserRoleBatchAddByRoleCode(tx *mysqlx.MysqlClient, userId string, roleCodes []string) error {
	return getTx(tx).Transaction(func(newTx *mysqlx.MysqlClient) error {
		list, err := RoleListByCodes(nil, roleCodes)
		if err != nil {
			return err
		}
		mapRoles := make(map[string]string, 0)
		for _, role := range list {
			if _, has := mapRoles[role.Code]; has {
				continue
			}
			mapRoles[role.Code] = role.Code
			err = UserRoleAdd(nil, &UserRoleModel{UserId: userId, RoleId: role.Id})
			if err != nil {
				return err
			}

		}
		return nil
	})
}
func UserRoleBatchDelByRoleCode(tx *mysqlx.MysqlClient, userId string, roleCodes []string) error {
	return getTx(tx).Transaction(func(newTx *mysqlx.MysqlClient) error {
		list, err := RoleListByCodes(newTx, roleCodes)
		if err != nil {
			return err
		}
		mapRoles := make(map[string]string, 0)
		for _, role := range list {
			if _, has := mapRoles[role.Code]; has {
				continue
			}
			mapRoles[role.Code] = role.Code
			err = UserRoleDel(newTx, userId, role.Id)
			if err != nil {
				return err
			}

		}
		return nil
	})
}
