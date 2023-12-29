package define

import (
	"encoding/json"
	"fmt"
)

const (
	NOTIFY_TYPE_EMAIL   = "email"
	NOTIFY_TYPE_WEBHOOK = "webhook"
)

type NotifyConfigEmail struct {
	Addr     string `json:"addr"`     //服务地址
	Port     int32  `json:"port"`     //服务端口
	Ssl      bool   `json:"ssl"`      //是否使用ssl
	Sender   string `json:"sender"`   //发件人
	Password string `json:"password"` //密码
}

type NotifyConfigWebhookHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type NotifyConfigWebhook struct {
	Url     string                       `json:"url"` //请求地址
	Headers []*NotifyConfigWebhookHeader `json:"headers"`
}

type NotifyTemplateEmail struct {
	Title    string `json:"title"`     //标题
	Receiver string `json:"receivers"` //接收者用;隔开
	Content  string `json:"content"`   //模板内容 ${name}替换
}

type NotifyTemplateWebhook struct {
	IsCustom bool   `json:"is_custom"` //default custom
	Content  string `json:"content"`
}

func IsNotifyConfigDefValid(notifyType string, configDef string) (any, error) {
	if notifyType == NOTIFY_TYPE_EMAIL {
		ret := &NotifyConfigEmail{}
		err := json.Unmarshal([]byte(configDef), ret)
		if err != nil {
			return nil, err
		}
		if ret.Addr == "" || ret.Port == 0 || ret.Sender == "" || ret.Password == "" {
			return nil, fmt.Errorf("格式化参数错误")
		}
		return ret, nil
	} else if notifyType == NOTIFY_TYPE_WEBHOOK {
		ret := &NotifyConfigWebhook{}
		err := json.Unmarshal([]byte(configDef), ret)
		if err != nil {
			return nil, err
		}
		if ret.Url == "" || ret.Headers == nil {
			return nil, fmt.Errorf("格式化参数错误")
		}
		return ret, nil
	} else {
		return nil, fmt.Errorf("不支持该通知类型")
	}
}

func IsNotifyTemplateDefValid(notifyType string, templateDef string) (any, error) {
	if notifyType == NOTIFY_TYPE_EMAIL {
		ret := &NotifyTemplateEmail{}
		err := json.Unmarshal([]byte(templateDef), ret)
		if err != nil {
			return nil, err
		}
		if ret.Title == "" || ret.Receiver == "" || ret.Content == "" {
			return nil, fmt.Errorf("格式化参数错误")
		}
		return ret, nil
	} else if notifyType == NOTIFY_TYPE_WEBHOOK {
		ret := &NotifyTemplateWebhook{}
		err := json.Unmarshal([]byte(templateDef), ret)
		if err != nil {
			return nil, err
		}

		if ret.IsCustom {
			if ret.Content == "" {
				return nil, fmt.Errorf("自定义内容必填")
			}
		}
		return ret, nil
	} else {
		return nil, fmt.Errorf("不支持该通知类型")
	}
}
