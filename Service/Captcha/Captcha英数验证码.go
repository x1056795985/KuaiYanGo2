package Captcha

import (
	"github.com/mojocn/base64Captcha"
	"server/global"
	"time"
)

const CAPTCHA = "captcha:"

type H缓存验证码 struct {
}

// 置验证码缓存  有效5分钟
func (r H缓存验证码) Set(id string, value string) {
	key := CAPTCHA + id
	global.H缓存.Set(key, value, time.Minute*5)
}

// 取验证码缓存
func (r H缓存验证码) Get(id string, 是否删除 bool) string {
	key := CAPTCHA + id

	val, ok := global.H缓存.Get(key)
	if !ok {
		return ""
	}
	if 是否删除 {
		global.H缓存.Delete(key)
	}
	return val.(string)
}

// 校验验证码  就是多一步 校验,实际没区别
func (r H缓存验证码) Verify(id, Value string, 是否删除 bool) bool {
	if id == "" || Value == "" {
		return false
	}
	v := H缓存验证码{}.Get(id, 是否删除)
	if v == Value { //如果验证成功 直接删除,防止验证码多次使用
		global.H缓存.Delete(CAPTCHA + id)
	}
	return v == Value
}

func Captcha_取英数验证码() (验证码id, Base64验证码图片 string, err error) {
	// 字符,公式,验证码配置
	// 生成默认数字的 driver

	driver := base64Captcha.NewDriverDigit(global.GVA_CONFIG.Captcha.ImgHeight, global.GVA_CONFIG.Captcha.ImgWidth, global.GVA_CONFIG.Captcha.KeyLong, 0.7, 80)
	// cp := base64Captcha.NewCaptcha(driver, store.UseWithCtx(c))   // v8下使用redis
	cp := base64Captcha.NewCaptcha(driver, H缓存验证码{})

	return cp.Generate()
}
