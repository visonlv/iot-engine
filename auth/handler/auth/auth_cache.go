package auth

import (
	"encoding/json"
	"fmt"
	"math"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/iot-engine/auth/model"
	"github.com/visonlv/iot-engine/auth/rediskey"
)

var AuthObj = NewAuth()

type Rule struct {
	Name      string `json:"name"`       //规则名称
	ValueType string `json:"value_type"` //数值类型
	Value     string `json:"value"`      //数值
}

type Permission struct {
	RoleCode string           //角色代码
	AppCode  string           //应用代码
	Urls     map[string]bool  //URL-bol
	Rules    map[string]*Rule //rulename-rule
	Menu     string           //JSON菜单
}
type RolePermission struct {
	Permissions map[string]*Permission //appcode-权限
	GenericUrls map[string]bool        //通用API权限
}
type Auth struct {
	WhiteList       *sync.Map //白名单url列表
	RolePermissions *sync.Map // 每个角色对应的ApiId列表 [角色代码]=角色权限信息
}

func NewAuth() *Auth {
	return &Auth{
		RolePermissions: new(sync.Map),
		WhiteList:       new(sync.Map),
	}
}

func (a *Auth) Start() {
	err := a.ResetAll()
	if err != nil {
		panic(err)
	}
	err = a.ResetWhiteList()
	if err != nil {
		panic(err)
	}
	go a.DelTimeOutLoginRecord()
}

func (a *Auth) ResetAll() error {
	list, err := model.PermissionList(nil)
	if err != nil {
		return err
	}
	resources, _, err := model.ResourcePage(nil, 1, math.MaxInt32, "", "", "")
	if err != nil {
		return err
	}
	mapResources := make(map[string]*model.ResourceModel)
	for _, res := range resources {
		mapResources[res.Id] = res
	}
	code2Info := make(map[string]*RolePermission, 0)
	for _, item := range list {
		cc, ok := code2Info[item.RoleCode]
		if !ok {
			cc = &RolePermission{
				Permissions: make(map[string]*Permission),
				GenericUrls: make(map[string]bool),
			}
			// cc = &Permission{
			// 	RoleCode: item.RoleCode,
			// 	AppCode:  item.AppCode,
			// 	Urls:     make(map[string]bool),
			// 	Rules:    make(map[string]*Rule),
			// }
			code2Info[item.RoleCode] = cc
		}
		per, has := cc.Permissions[item.AppCode]
		if item.AppCode != "" && !has {
			per = &Permission{
				RoleCode: item.RoleCode,
				AppCode:  item.AppCode,
				Urls:     make(map[string]bool),
				Rules:    make(map[string]*Rule),
			}
			cc.Permissions[item.AppCode] = per
		}
		//解析资源内容
		resItems := []*model.PermissionResource{}
		err := json.Unmarshal([]byte(item.Resources), &resItems)
		if err != nil {
			return err
		}
		for _, resItem := range resItems {
			res, has := mapResources[resItem.ResourceId]
			if !has {
				fmt.Println(item.AppCode, item.RoleCode, item.Resources, "不存在")
				continue
			}
			if item.AppCode == "" {
				cc.GenericUrls[res.Content] = true
				continue
			}
			switch res.Type {
			case model.ResourceTypeApi:
				per.Urls[res.Content] = true
			case model.ResourceTypeMenu:
				per.Menu = res.Content
			case model.ResourceTypeRule:
				{
					//解析规则内容
					ruleItem := &Rule{Name: res.Name, ValueType: res.Content, Value: resItem.Value}
					per.Rules[ruleItem.Name] = ruleItem
				}
			default:
			}
		}
	}
	bb, _ := json.Marshal(code2Info)
	logger.Infof("[auth] ResetAll code2Info:%s", string(bb))
	temp := new(sync.Map)
	for code, v := range code2Info {
		temp.Store(code, v)
	}
	a.RolePermissions = temp
	return nil
}
func (a *Auth) ResetWhiteList() error {
	//清空地址
	wl := new(sync.Map)
	list, _, err := model.ApiWhiteListPage(nil, 1, math.MaxInt32, "")
	if err != nil {
		return err
	}
	for _, item := range list {
		wl.Store(item.Path, "")
	}
	a.WhiteList = wl
	return nil
}

func (a *Auth) IsWhite(url string) bool {
	//判断白名单
	isWhite := false
	a.WhiteList.Range(func(key, value interface{}) bool {
		b, err := path.Match(key.(string), url)
		if err != nil {
			logger.Errorf("[auth] IsWhite match url:%s err:%s", url, err)
			return false
		}
		if b {
			isWhite = true
			return false
		}
		return true
	})
	return isWhite
}

func (a *Auth) IsPemission(appcode string, roleList []string, url string) bool {
	//取出三种情况的可能匹对的地址
	matchUrls := []string{url}
	//第二种情况：/rpc/xxx/xxxxService.* 是否匹配
	index := strings.LastIndex(url, ".")
	if index > 0 {
		matchUrls = append(matchUrls, url[:index+1]+"*")
	}
	//第三种情况：/rpc/xxxx/*  是否匹配
	index = strings.LastIndex(url, "/")
	if index > 0 {
		matchUrls = append(matchUrls, url[:index+1]+"*")
	}

	//判断各个角色权限 匹配一个则满足
	for _, role := range roleList {
		rp, ok := a.RolePermissions.Load(role)
		if ok {

			per, has := rp.(*RolePermission).Permissions[appcode]
			for _, u := range matchUrls {
				//查询通用api
				if _, b := rp.(*RolePermission).GenericUrls[u]; b {
					return true
				}
				//查询应用的接口权限
				if has {
					if _, b := per.Urls[u]; b {
						return true
					}
				}

			}

		}

	}
	return false
}

func (a *Auth) DelTimeOutLoginRecord() {
	tm := rediskey.TokenKey.Expire
	ticker := time.NewTicker(tm)
	for {
		//清理过期的登录数据
		model.LoginRecordDelByUpdateTime(nil, time.Now().Add(-1*tm).Format(model.TIMETEMPLATE))
		<-ticker.C
	}
}

func (a *Auth) IsAllowLogin(appcode, rolecode string) bool {
	rp, ok := a.RolePermissions.Load(rolecode)
	if !ok {
		return false
	}
	per, has := rp.(*RolePermission).Permissions[appcode]
	if !has {
		return false
	}
	_, ok = per.Rules[model.RuleNameLogin]
	if ok {
		return true
	}
	return false
}
func (a *Auth) IsAllowMultiSignOn(appcode, rolecode string) bool {

	rp, ok := a.RolePermissions.Load(rolecode)
	if !ok {
		return false
	}
	per, has := rp.(*RolePermission).Permissions[appcode]
	if !has {
		return false
	}
	_, ok = per.Rules[model.RuleNameMultiSingOn]
	if ok {
		return true
	}
	return false
}

func RoleCodesByAppCode(appcode string) []string {
	roles := []string{}
	AuthObj.RolePermissions.Range(func(key, value interface{}) bool {
		rp := value.(*RolePermission)
		if per, has := rp.Permissions[appcode]; has {
			if _, ok := per.Rules[model.RuleNameLogin]; ok {
				roles = append(roles, per.RoleCode)
			}
		}
		return true
	})
	return roles
}
