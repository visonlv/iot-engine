package verification

import (
	"errors"
	"fmt"
	"time"

	"github.com/visonlv/go-vkit/logger"
	"github.com/visonlv/iot-engine/auth/app"
	"github.com/visonlv/iot-engine/auth/model"
	"github.com/visonlv/iot-engine/auth/rediskey"
	"github.com/visonlv/iot-engine/auth/utils"
	"gorm.io/gorm"
)

type NoticeMessageReq struct {
	Code   string `json:"code"`
	ErrMsg string `json:"errMsg"`
}
type ErrMsg struct {
	Code     string `json:"code"`
	MsgId    string `json:"msgId"`
	Time     string `json:"time"`
	ErrorMsg string `json:"errorMsg"`
}

func VerificationEmail(email string) error {
	m, err := model.UserGetByEmail(nil, email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("邮箱不存在")
		}
		return err
	}
	code, err := app.Redis.GetString(rediskey.SendMailIntervalKey, email)
	if err == nil {
		return fmt.Errorf("获取校验码太频繁 间隔:%ds", rediskey.SendMailIntervalKey.Expire/time.Second)
	}
	code = utils.RandomCount(6)
	content := fmt.Sprintf(`
    <p> 您好 %s,</p>
		<p style="text-indent:2em">您在语音平台申请找回密码，验证码:%s</p> 
	`, m.NickName, code)

	err = utils.SendMail(email, "语音平台找回密码", content)
	if err != nil {
		logger.Errorf("发送邮件失败 %s %s", email, err.Error())
		return fmt.Errorf("发送邮件失败 %s", email)
	}
	err = app.Redis.Set(rediskey.SendMailIntervalKey, email, code)
	if err != nil {
		return fmt.Errorf("设置缓存失败 %s", err.Error())
	}

	err = app.Redis.Set(rediskey.VerificationCodeKey, email, code)
	if err != nil {
		return fmt.Errorf("设置缓存失败 %s", err.Error())
	}
	return nil
}
func VerificationPhone(appcode, phone string) error {
	_, err := model.UserGetByPhone(nil, phone)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("手机号未注册")
		}
		return err
	}
	rKey := appcode + "_" + phone
	_, err = app.Redis.GetString(rediskey.SendMessageIntervalKey, rKey)
	if err == nil {
		return fmt.Errorf("获取验证码太频繁 间隔:%ds", rediskey.SendMessageIntervalKey.Expire/time.Second)
	}
	code := utils.RandomCount(4)
	message := fmt.Sprintf(`您的验证码是%s。`, code)

	if app.Cfg.Env == "dev" {
		err = SendPhoneMessageByHttp(phone, message)
	} else {
		err = SendPhoneMessageByRpc(phone, message)
	}
	if err != nil {
		logger.Errorf("发送短信失败 %s %s", phone, err.Error())
		return fmt.Errorf("发送短信失败")
	}
	err = app.Redis.Set(rediskey.SendMessageIntervalKey, rKey, code)
	if err != nil {
		return fmt.Errorf("设置缓存失败 %s", err.Error())
	}

	err = app.Redis.Set(rediskey.PhoneMessageCodeKey, rKey, code)
	if err != nil {
		return fmt.Errorf("设置缓存失败 %s", err.Error())
	}
	return nil
}
func SendPhoneMessageByRpc(phone, message string) error {
	return nil
}
func SendPhoneMessageByHttp(phone, message string) error {

	return nil
}
