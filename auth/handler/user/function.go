package user

import (
	"fmt"

	"github.com/visonlv/go-vkit/utilsx"
	"github.com/visonlv/iot-engine/auth/handler/auth"
	"github.com/visonlv/iot-engine/auth/model"
)

var passwordKey = "infore_2022~"

func EncodePassword(password string) string {
	return utilsx.Md5Encode(fmt.Sprintf("%s%s", passwordKey, password))
}

func RolesByUserAndAppCode(userId string, appCode string) ([]string, error) {
	appRoles := auth.RoleCodesByAppCode(appCode)
	mapRoleCodes := make(map[string]string)
	for _, rolecode := range appRoles {
		mapRoleCodes[rolecode] = appCode
	}
	roles, err := model.RoleListByUserId(nil, userId)
	if err != nil {
		return nil, err
	}
	roleCodes := []string{}
	for _, rr := range roles {
		if _, has := mapRoleCodes[rr.Code]; has {
			roleCodes = append(roleCodes, rr.Code)
		}
	}
	return roleCodes, nil
}
