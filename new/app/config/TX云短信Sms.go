package config

import (
	"EFunc/utils"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"server/config"
	"server/global"
	"server/new/app/service"
	"time"
)

func TX云短信Sms() config.TX云短信Sms {
	var value = config.TX云短信Sms{}

	temp, ok := global.H缓存.Get("config.TX云短信Sms")
	if ok {
		value, ok = temp.(config.TX云短信Sms)
		if ok {
			return value
		}
	}
	局_计时 := utils.S时间_取现行时间戳13()
	setting := service.S_newSetting(global.GVA_DB)
	c := gin.Context{}
	str, err := setting.Info(&c, "config.TX云短信Sms")
	if err == nil {
		err = json.Unmarshal([]byte(str), &value)
		if err == nil && utils.S时间_取现行时间戳13()-局_计时 > 500 { //超过100耗秒自动缓存5分钟
			global.H缓存.Set("config.TX云短信Sms", value, 3000*time.Second)
		}
	}

	return value
}
