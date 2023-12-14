package convert

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/visonlv/iot-engine/auth/handler/user"
	pb "github.com/visonlv/iot-engine/auth/proto"

	"github.com/visonlv/iot-engine/auth/model"

	"github.com/visonlv/go-vkit/utilsx"
)

func UserToUserPb(m *model.UserModel, appCode string) (*pb.User, error) {
	itemRet := &pb.User{}
	utilsx.DeepCopy(m, itemRet)
	itemRet.Password = ""
	itemRet.CreateTime = m.CreatedAt.UnixMilli()
	itemRet.CreateUsername = model.UserNameFromCache(nil, m.CreateUser)
	if appCode != "" {
		//查询对应的角色
		codes, err := user.RolesByUserAndAppCode(m.Id, appCode)
		if err != nil {
			return nil, err
		}
		itemRet.RoleCode = codes
	} else {
		codes, err := model.RoleCodeListByUserId(nil, m.Id)
		if err != nil {
			return nil, err
		}
		itemRet.RoleCode = codes
	}

	return itemRet, nil
}
func PbRoleToRole(m *pb.Role) *model.RoleModel {
	itemRet := &model.RoleModel{}
	utilsx.DeepCopy(m, itemRet)
	itemRet.Rolename = m.Name
	itemRet.CreatedAt = time.Now()
	itemRet.CreateUser = m.CreateUser
	return itemRet
}
func RoleToPbRole(m *model.RoleModel) (*pb.Role, error) {
	itemRet := &pb.Role{}
	utilsx.DeepCopy(m, itemRet)
	itemRet.Name = m.Rolename
	itemRet.CreateTime = m.CreatedAt.UnixMilli()
	itemRet.CreateUsername = model.UserNameFromCache(nil, m.CreateUser)
	return itemRet, nil
}

func ResourceStringsToJsons(resources string, resource_type string) (string, error) {
	list := []*model.PermissionResource{}
	err := json.Unmarshal([]byte(resources), &list)
	if err != nil {
		return "", err
	}
	mapPer := make(map[string]string)
	for _, per := range list {
		mapPer[per.ResourceId] = per.Value
	}
	ids := []string{}
	for _, res := range list {
		ids = append(ids, res.ResourceId)
	}
	items, err := model.ResourceListByIds(nil, ids)
	if err != nil {
		return "", err
	}
	switch resource_type {
	case model.ResourceTypeApi:
		stru := []interface{}{}
		for _, item := range items {
			jsonMap := make(map[string]string)
			jsonMap["api"] = item.Content
			jsonMap["name"] = item.Name
			stru = append(stru, struct {
				Name string `json:"name,omitempty"`
				Api  string `json:"api,omitempty"`
			}{item.Name, item.Content})
		}
		js, err := json.Marshal(stru)
		return string(js), err
	case model.ResourceTypeMenu:
		menus := make([]interface{}, 0)
		for _, item := range items {
			temp := make([]interface{}, 0)
			err := json.Unmarshal([]byte(item.Content), &temp)
			if err != nil {
				return "", err
			}
			menus = append(menus, temp...)
		}
		bbb, _ := json.Marshal(menus)
		return string(bbb), nil
	case model.ResourceTypeRule:
		stru := []interface{}{}
		for _, item := range items {
			jsonMap := make(map[string]string)
			jsonMap["name"] = item.Name
			jsonMap["value_type"] = item.Content
			jsonMap["value"] = mapPer[item.Id]
			stru = append(stru, struct {
				Name      string `json:"name,omitempty"`
				ValueType string `json:"value_type,omitempty"`
				Value     string `json:"value,omitempty"`
			}{item.Name, item.Content, mapPer[item.Id]})
		}
		js, err := json.Marshal(stru)
		return string(js), err
	}
	return "", errors.New(resource_type + " 未知的资源类型")
}
