package model

import (
	"fmt"
	"time"

	"github.com/visonlv/go-vkit/mysqlx"
	"github.com/visonlv/go-vkit/utilsx"
	"gorm.io/gorm"
)

// 权限表
var (
	permission_model = &PermissionModel{}
)

// 结构体
type PermissionModel struct {
	Id         string    `gorm:"primaryKey;type:varchar(64);comment:主键id"`
	CreatedAt  time.Time // 创建时间
	UpdatedAt  time.Time // 更新时间
	CreateUser string    `gorm:"type:varchar(128);comment:创建用户"`
	UpdateUser string    `gorm:"type:varchar(128);comment:更新用户"`
	IsDelete   int       `gorm:"type:tinyint;not null;default:0;comment:删除状态 0正常 1删除"`

	RoleId       string `gorm:"type:varchar(64);comment:角色ID;"`
	AppId        string `gorm:"type:varchar(64);comment:应用ID;"`
	ResourceType string `gorm:"type:varchar(128);comment:资源类型(api/menu/rule)"`
	Resources    string `gorm:"type:text;comment:资源内容"`
}

func (a *PermissionModel) BeforeCreate(tx *gorm.DB) error {
	a.Id = utilsx.GenUuid()
	return nil
}

func (*PermissionModel) TableName() string {
	return "t_permission"
}
func PermissionPage(tx *mysqlx.MysqlClient, pageIndex int32, pageSize int32) ([]*PermissionModel, int32, error) {
	items := []*PermissionModel{}
	var total int32
	err := getTx(tx).Model(permission_model).Where("is_delete = ?", 0).Order("created_at desc,id").FindPage(pageIndex, pageSize, items, &total)
	return items, total, err
}
func PermissionList(tx *mysqlx.MysqlClient) ([]*RoleAppPermission, error) {
	items := []*RoleAppPermission{}
	err := getTx(tx).Raw("select r.code as role_code,a.code as app_code,p.resource_type,p.resources from t_permission p left join t_role r on p.role_id=r.id left join t_app a on a.id=p.app_id where p.is_delete = 0 order by p.created_at desc").FindList(&items)
	return items, err
}
func PermissionInfoPage(tx *mysqlx.MysqlClient, pageIndex int32, pageSize int32, roleId, appId string) ([]*PermissionInfo, int32, error) {
	items := []*PermissionInfo{}
	var total int32
	sql := "select p.id,p.role_id,r.rolename as role_name ,p.app_id,a.name as app_name,p.resource_type ,p.created_at,p.resources from t_permission p left join t_role r on p.role_id=r.id left join t_app a on p.app_id=a.id where p.is_delete = 0"
	if roleId != "" {
		sql += fmt.Sprintf(" AND (p.role_id = '%s' or p.role_id = '')", roleId)
	}

	if appId != "" {
		sql += fmt.Sprintf(" AND (p.app_id='%s' or p.app_id = '')", appId)
	}
	sql += " order by p.created_at desc,r.rolename"

	err := getTx(tx).FindRawPage(sql, pageIndex, pageSize, &items, &total)
	return items, total, err
}

func PermissionGetByInfo(tx *mysqlx.MysqlClient, appId, roleId, resType string) (*PermissionModel, bool, error) {
	item := &PermissionModel{}
	exist, err := getTx(tx).Where("app_id = ? AND role_id = ? AND resource_type = ? AND is_delete = ?", appId, roleId, resType, 0).FindFirst(item)
	return item, exist, err
}

// 添加权限
func PermissionAdd(tx *mysqlx.MysqlClient, m *PermissionModel) error {
	if err := getTx(tx).Model(permission_model).Insert(m); err != nil {
		return err
	}
	return nil
}
func PermissionGet(tx *mysqlx.MysqlClient, id string) (*PermissionModel, error) {
	item := &PermissionModel{}
	result := getTx(tx).Where("id = ? AND is_delete = ?", id, 0).First(item)
	return item, result.GetDB().Error
}
func PermissionDel(tx *mysqlx.MysqlClient, id string) error {
	result := getTx(tx).Model(permission_model).Where("id = ?", id).Update("is_delete", 1)
	return result.GetDB().Error
}
func PermissionUpdate(tx *mysqlx.MysqlClient, m *PermissionModel) (*PermissionModel, error) {
	err := getTx(tx).UpdateEx(m)
	return m, err
}
func PermissionExistByResourceId(tx *mysqlx.MysqlClient, resourceId string) (bool, error) {
	return getTx(tx).Raw(fmt.Sprintf("SELECT * FROM t_permission where is_delete = 0 AND resources like '%%\"resource_id\":\"%s\"%%'", resourceId)).FindFirst(&PermissionModel{})
}

func PermissionGetByRoleIdAndAppIdAndResourceType(tx *mysqlx.MysqlClient, roleId, appId, resourceType string) (*PermissionModel, error) {
	item := &PermissionModel{}
	result := getTx(tx).Where("role_id = ? AND app_id = ? AND resource_type = ? AND is_delete = ?", roleId, appId, resourceType, 0).First(item)
	return item, result.GetDB().Error
}

func PermissionCountByRoleId(tx *mysqlx.MysqlClient, roleId string) (count int64, err error) {
	result := getTx(tx).Model(permission_model).Where("role_id = ? AND is_delete = ?", roleId, 0).Count(&count)
	err = result.GetDB().Error
	return
}

func PermissionCountByAppId(tx *mysqlx.MysqlClient, appId string) (count int64, err error) {
	result := getTx(tx).Model(permission_model).Where("app_id = ? AND is_delete = ?", appId, 0).Count(&count)
	err = result.GetDB().Error
	return
}
