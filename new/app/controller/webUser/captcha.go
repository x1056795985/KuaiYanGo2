package controller

import (
	"github.com/gin-gonic/gin"
	"server/Service/Captcha"
	"server/global"
	"server/new/app/controller/Common/response"
	"server/new/app/models/constant"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	"time"
)

func (b *Base) Captcha2(c *gin.Context) {
	var 请求 struct {
		AppId int `json:"AppId" binging:"required,min=10000"` // Appid 必填
	}
	//解析失败
	if !b.ToJSON(c, &请求) {
		return
	}
	var info = struct {
		网页用户中心配置 dbm.DB_AppInfoWebUser
	}{}
	var err error
	tx := *global.GVA_DB

	info.网页用户中心配置, err = service.NewAppInfoWebUser(c, &tx).Info(请求.AppId)
	if err != nil || info.网页用户中心配置.Status != 1 {
		response.FailWithMessage(c, constant.C常_关闭提示)
		return
	}
	// 判断验证码是否开启
	openCaptcha := info.网页用户中心配置.CaptchaLogin                          // 是否开启防爆次数
	openCaptchaTimeOut := global.GVA_CONFIG.Captcha.OpenCaptchaTimeOut // 缓存超时时间
	key := c.ClientIP()                                                //获取客户端ip
	v, ok := global.H缓存.Get(key)                                       //获取
	if !ok {
		global.H缓存.Set(key, 1, time.Second*time.Duration(openCaptchaTimeOut))
	}

	var oc bool
	if openCaptcha == 0 || openCaptcha < interfaceToInt(v) {
		oc = true
	}

	验证码id, Base64验证码图片, err := Captcha.Captcha_取点选验证码(interfaceToInt(v))
	if err != nil {
		response.FailWithMessage(c, "验证码获取失败")
		return
	}

	response.OkWithDetailed(c, sysCaptchaResponse{
		CaptchaId:     验证码id,
		PicPath:       Base64验证码图片,
		CaptchaLength: global.GVA_CONFIG.Captcha.KeyLong,
		OpenCaptcha:   oc,
	}, "ok")

}

// 验证码api  data
type sysCaptchaResponse struct {
	CaptchaId     string `json:"captchaId"`     //验证码id
	PicPath       string `json:"picPath"`       //验证码数据
	CaptchaLength int    `json:"captchaLength"` //验证码长度
	OpenCaptcha   bool   `json:"openCaptcha"`   //是否显示验证码
}

// 类型转换
func interfaceToInt(v interface{}) (i int) {
	switch v := v.(type) {
	case int:
		i = v
	default:
		i = 0
	}
	return
}
