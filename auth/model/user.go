package model

import (
	"errors"
	"fmt"
	"time"

	"github.com/visonlv/iot-engine/auth/app"

	"github.com/visonlv/iot-engine/auth/rediskey"

	"github.com/go-redis/redis/v8"
	"github.com/visonlv/go-vkit/mysqlx"
	"github.com/visonlv/go-vkit/utilsx"
	"gorm.io/gorm"
)

// 用户
var (
	user_model = &UserModel{}
)

type UserModel struct {
	Id         string    `gorm:"primaryKey;type:varchar(64);comment:主键id"`
	CreatedAt  time.Time // 创建时间
	UpdatedAt  time.Time // 更新时间
	CreateUser string    `gorm:"type:varchar(128);comment:创建用户"`
	UpdateUser string    `gorm:"type:varchar(128);comment:更新用户"`
	IsDelete   int       `gorm:"type:tinyint;not null;default:0;comment:删除状态 0正常 1删除"`

	Account  string `gorm:"type:varchar(128);comment:账号"`
	Password string `gorm:"type:varchar(128);comment:密码"`
	NickName string `gorm:"type:varchar(128);comment:用户昵称"`
	Email    string `gorm:"type:varchar(64);comment:邮箱"`
	Phone    string `gorm:"type:varchar(64);comment:手机号码"`
}

func (a *UserModel) BeforeCreate(tx *gorm.DB) error {
	if a.Id == "" {
		a.Id = utilsx.GenUuid()
	}
	return nil
}

func (*UserModel) TableName() string {
	return "t_user"
}

func UserAddWithTransaction(tx *mysqlx.MysqlClient, m *UserModel, roles []*RoleModel) error {
	return getTx(tx).Transaction(func(newTx *mysqlx.MysqlClient) error {

		//邮箱、手机号和账号不可重复
		if len(m.Email) != 0 {
			_, err := UserGetByEmail(newTx, m.Email)
			if err == nil {
				return errors.New("邮箱已存在")
			}
		}
		if len(m.Account) != 0 {
			_, err := UserGetByAccount(newTx, m.Account)
			if err == nil {
				return errors.New("账号已存在")
			}
		}
		if len(m.Phone) != 0 {
			_, err := UserGetByPhone(newTx, m.Phone)
			if err == nil {
				return errors.New("手机号已存在")
			}
		}
		if err := UserAdd(newTx, m); err != nil {
			return err
		}
		//添加新的关系
		for _, v := range roles {
			if err := UserRoleAdd(newTx, &UserRoleModel{
				UserId: m.Id,
				RoleId: v.Id,
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func UserDelWithTransaction(tx *mysqlx.MysqlClient, userId string) error {
	return getTx(tx).Transaction(func(newTx *mysqlx.MysqlClient) error {
		if err := UserDel(newTx, userId); err != nil {
			return err
		}
		//删除原有关系
		if err := UserRoleDelByUserId(newTx, userId); err != nil {
			return err
		}
		return nil
	})
}

func UserGetFromCache(tx *mysqlx.MysqlClient, id string) (*UserModel, error) {
	to := &UserModel{}
	err := app.Redis.GetJson(rediskey.UserCacheKey, id, to)
	if err != nil && err != redis.Nil {
		return nil, err
	}
	if err == nil {
		if to.Id == "" {
			return nil, fmt.Errorf("数据不存在")
		}
		return to, nil
	} else {
		newOne, err := UserGet(nil, id)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				app.Redis.SetJson(rediskey.UserCacheKey, id, to)
			}
			return nil, err
		}
		app.Redis.SetJson(rediskey.UserCacheKey, id, newOne)
		return newOne, nil
	}
}

func UserNameFromCache(tx *mysqlx.MysqlClient, id string) string {
	if id == "" {
		return "-"
	}
	createU, err := UserGetFromCache(nil, id)
	if err == nil {
		return createU.NickName
	} else {
		return "-"
	}
}

func UserAdd(tx *mysqlx.MysqlClient, m *UserModel) error {
	if err := getTx(tx).Model(user_model).Insert(m); err != nil {
		return err
	}
	return nil
}

func UserDel(tx *mysqlx.MysqlClient, id string) error {
	result := getTx(tx).Model(user_model).Where("id = ?", id).Update("is_delete", 1)
	if result.GetDB().Error == nil {
		app.Redis.Del(rediskey.UserCacheKey, id)
	}
	return result.GetDB().Error
}

func UserListByCreateUser(tx *mysqlx.MysqlClient, id string) (list []*UserModel, err error) {
	err = getTx(tx).Model(user_model).Where("create_user = ? AND is_delete = ?", id, 0).FindList(&list)
	return
}

func UserGet(tx *mysqlx.MysqlClient, id string) (*UserModel, error) {
	item := &UserModel{}
	result := getTx(tx).Where("id = ? AND is_delete = ?", id, 0).First(item)
	return item, result.GetDB().Error
}

func UserGetInIds(tx *mysqlx.MysqlClient, ids []string) (list []*UserModel, err error) {
	err = getTx(tx).Where("id in ? AND is_delete = ?", ids, 0).FindList(&list)
	return
}

func UserGetByEmail(tx *mysqlx.MysqlClient, email string) (*UserModel, error) {
	item := &UserModel{}
	result := getTx(tx).Where("email = ? AND is_delete = ?", email, 0).First(item)
	return item, result.GetDB().Error
}
func UserGetByAccount(tx *mysqlx.MysqlClient, account string) (*UserModel, error) {
	item := &UserModel{}
	result := getTx(tx).Where("account = ? AND is_delete = ?", account, 0).First(item)
	return item, result.GetDB().Error
}
func UserEmailExists(tx *mysqlx.MysqlClient, email string) (bool, error) {
	item := &UserModel{}
	has, err := getTx(tx).Where("email = ? AND is_delete = ?", email, 0).Exists(item)
	return has, err
}

func UserUpdate(tx *mysqlx.MysqlClient, m *UserModel) (*UserModel, error) {
	err := getTx(tx).UpdateEx(m)
	if err == nil {
		app.Redis.Del(rediskey.UserCacheKey, m.Id)
	}
	return m, err
}

func UserCountByCreateUserId(tx *mysqlx.MysqlClient, userId string) (count int64, err error) {
	result := getTx(tx).Model(user_model).Where("is_delete = ? AND create_user=?", 0, userId).Count(&count)
	return count, result.GetDB().Error
}
func UserGetAndDelete(tx *mysqlx.MysqlClient, id string) (*UserModel, error) {
	item := &UserModel{}
	result := getTx(tx).Where("id = ?", id).First(item)
	return item, result.GetDB().Error
}

func UserUpdateAndRole(tx *mysqlx.MysqlClient, u *UserModel, roleCodes []string) error {
	return getTx(tx).Transaction(func(newTx *mysqlx.MysqlClient) error {
		_, err := UserUpdate(newTx, u)
		if err != nil {
			return err
		}
		//删除角色
		err = UserRoleDelByUserId(newTx, u.Id)
		if err != nil {
			return err
		}
		//重新配置角色关联
		if len(roleCodes) != 0 {
			roles, err := RoleListByCodes(newTx, roleCodes)
			if err != nil {
				return err
			}
			for _, role := range roles {
				err = UserRoleAdd(newTx, &UserRoleModel{
					UserId: u.Id, RoleId: role.Id,
				})
				if err != nil {
					return err
				}
			}

		}
		return nil
	})
}
func UserPageBySql(tx *mysqlx.MysqlClient, filter string, pageIndex int32, pageSize int32, orderKey string, desc bool) (list []*UserModel, total int32, err error) {
	query := getTx(tx).Model(user_model).Where("is_delete=0")
	if filter != "" {
		query = query.Where(filter)
	}
	if len(orderKey) > 0 {
		if desc {
			query = query.Order(orderKey + " desc")
		} else {
			query = query.Order(orderKey)
		}
	} else {
		query = query.Order("created_at desc")
	}
	err = query.FindPage(pageIndex, pageSize, &list, &total)
	return
}

func UserAddAndRole(tx *mysqlx.MysqlClient, u *UserModel, roleIds []string) error {
	return getTx(tx).Transaction(func(newTx *mysqlx.MysqlClient) error {
		err := UserAdd(newTx, u)
		if err != nil {
			return err
		}
		for _, id := range roleIds {
			err = UserRoleAdd(newTx, &UserRoleModel{UserId: u.Id, RoleId: id})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func UserPage(tx *mysqlx.MysqlClient, pageIndex int32, pageSize int32, orderKey string, desc bool, account, nickName, phone, email, userId string, roleCodes []string) (list []*UserModel, total int32, err error) {
	query := getTx(tx).Model(user_model).Where("is_delete = ?", 0)
	if len(roleCodes) > 0 {
		query = query.Where("id in (select user_id from t_user_role ur inner join t_role r on  r.code in ? and r.is_delete=0)", roleCodes)
	}

	if len(account) > 0 {
		query = query.Where("account like ?", "%"+account+"%")
	}
	if len(nickName) > 0 {
		query = query.Where("nick_name like ?", "%"+nickName+"%")
	}
	if len(email) > 0 {
		query = query.Where("email like ?", "%"+email+"%")
	}
	if len(phone) > 0 {
		query = query.Where("phone like ?", "%"+phone+"%")
	}
	if len(userId) > 0 {
		query = query.Where("create_user = ?", userId)
	}
	if len(orderKey) > 0 {
		if desc {
			query = query.Order(orderKey + " desc")
		} else {
			query = query.Order(orderKey)
		}
	} else {
		query = query.Order("created_at desc")
	}
	err = query.FindPage(pageIndex, pageSize, &list, &total)
	return
}

func UserList(tx *mysqlx.MysqlClient) (list []*UserModel, err error) {
	query := getTx(tx).Model(user_model)
	err = query.FindList(&list)
	return
}
func UserGetByPhone(tx *mysqlx.MysqlClient, phone string) (*UserModel, error) {
	item := &UserModel{}
	result := getTx(tx).Where("phone = ? AND is_delete = ?", phone, 0).First(item)
	return item, result.GetDB().Error
}
