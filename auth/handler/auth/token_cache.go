package auth

import (
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/iot-engine/auth/app"
	"github.com/visonlv/iot-engine/auth/model"
	"github.com/visonlv/iot-engine/auth/rediskey"
)

func AddToken(appcode, userId, roleCode, token string) error {
	err := app.Redis.Set(rediskey.TokenKey, token, userId)
	if err != nil {
		logger.Errorf("添加Token到Redis失败 %s", err.Error())
		return err
	}
	//是否多点登录
	isMultiSignOn := AuthObj.IsAllowMultiSignOn(appcode, roleCode)

	if !isMultiSignOn {
		items, err := model.LoginRecordListByUserAndRole(nil, userId, roleCode)
		if err != nil {
			logger.Errorf("查询登录记录失败 %s", err.Error())
			return err
		}
		if len(items) > 0 {
			//删除REDIS旧Token，更新新Token
			for _, item := range items {
				app.Redis.Del(rediskey.TokenKey, item.Token)
				model.LoginRecordDel(nil, item.Id)
			}
		}
	}
	//插入数据库
	m := &model.LoginRecordModel{
		AppCode:  appcode,
		UserId:   userId,
		RoleCode: roleCode,
		Token:    token,
	}
	err = model.LoginRecordAdd(nil, m)
	return err
}

func DeleteToken(token string) {
	app.Redis.Del(rediskey.TokenKey, token)
}

func UserTokenExist(token string) error {
	_, err := app.Redis.GetString(rediskey.TokenKey, token)
	if err == redis.Nil {
		return fmt.Errorf("账号在其他设备上登录，请确认!")
	}
	if err != nil {
		logger.Errorf("请求令牌失效，请求失败 %s", err.Error())
		return fmt.Errorf("请求令牌失效，请求失败!!")
	}
	return nil
}
