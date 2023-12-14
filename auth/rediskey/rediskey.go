package rediskey

import (
	"time"

	"github.com/visonlv/go-vkit/redisx"
)

var (
	TokenKey            *redisx.RedisKey = &redisx.RedisKey{Code: "iot-engine-auth:token", Expire: time.Hour * 24 * 7}
	UserCacheKey        *redisx.RedisKey = &redisx.RedisKey{Code: "iot-engine-auth:userInfo", Expire: time.Hour * 24}
	VerificationCodeKey *redisx.RedisKey = &redisx.RedisKey{Code: "iot-engine-auth:verificationCode", Expire: time.Minute * 5}
	SendMailIntervalKey *redisx.RedisKey = &redisx.RedisKey{Code: "iot-engine-auth:sendMailInterval", Expire: time.Second * 10}
	//手机短信验证码
	PhoneMessageCodeKey    *redisx.RedisKey = &redisx.RedisKey{Code: "iot-engine-auth:phoneMessageCode", Expire: time.Minute * 1}
	SendMessageIntervalKey *redisx.RedisKey = &redisx.RedisKey{Code: "iot-engine-auth:sendMessageInterval", Expire: time.Second * 10}
)
