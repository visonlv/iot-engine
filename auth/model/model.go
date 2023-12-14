package model

import (
	"time"

	"github.com/visonlv/iot-engine/auth/app"

	"github.com/visonlv/go-vkit/mysqlx"
)

const (
	ResourceTypeApi  = "api"
	ResourceTypeMenu = "menu"
	ResourceTypeRule = "rule"

	//时间模板
	TIMETEMPLATE = "2006-01-02 15:04:05"

	//手机方式获取验证
	VerificationMethodPhone = 1
	//邮箱方式获取验证
	VerificationMethodPEmail = 2

	//单点登录
	SingleSignOn = 0
	//多点登录
	MultiSignOn = 1

	//登录规则名称
	RuleNameLogin = "登录"
	//多点登录规则名称
	RuleNameMultiSingOn = "多点登录"
)

// InitTable 初始化数据库表
func InitTable() {
	// 自动建表
	app.Mysql.AutoMigrate(
		&AppModel{}, &ResourceModel{}, &UserRoleModel{}, &RoleModel{}, &UserModel{}, &ApiWhiteListModel{}, &LoginRecordModel{}, &PermissionModel{})
}

func getTx(tx *mysqlx.MysqlClient) *mysqlx.MysqlClient {
	if tx == nil {
		return app.Mysql
	}
	return tx
}

type RoleApi struct {
	AppCode  string
	RoleCode string
	Route    []string
}
type RoleAppPermission struct {
	AppCode      string
	RoleCode     string
	ResourceType string
	Resources    string
}

type PermissionResource struct {
	ResourceId string `json:"resource_id"`
	Value      string `json:"value"`
}

type PermissionInfo struct {
	Id           string    `json:"id,omitempty"`
	RoleId       string    `json:"role_id"`
	RoleName     string    `json:"role_name"`
	AppId        string    `json:"app_id,omitempty"`
	AppName      string    `json:"app_name,omitempty"`
	ResourceType string    `json:"resource_type,omitempty"`
	Resources    string    `json:"resources,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
}
