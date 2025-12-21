package middleware

import (
	. "EFunc/utils"
	"bytes"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"server/global"
	"server/new/app/controller/Common/response"
	"server/new/app/models/constant"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
	"time"
)

// Token有效的才放行,否则返回Ttoken失效
func IsTokenWebUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		局_白名单 := []string{
			"/userApi/app/getAppBaseInfo",
			"/userApi/app/getAppGongGao",
			"/userApi/base/loginUserOrKa",
			"/userApi/base/loginKey",
			"/userApi/user/newUserInfo",
			"/userApi/user/getPwSendSms",
			"/userApi/user/smsCodeSetPassWord",
			"/userApi/base/Captcha2",
		}
		if S数组_是否存在(局_白名单, c.Request.URL.Path) {
			c.Next()
			return
		}

		Token := c.Request.Header.Get("Token") //优先协议头的,Token
		if Token == "" {                       //如果协议头没有,再读取,url内的
			Token = c.Request.FormValue("Token")
		}
		if Token == "" { //如果协议头没有,再读取,cookies内的
			Token, _ = c.Cookie("Token")
		}

		if Token == "" {
			response.FailTokenErr(c, gin.H{"reload": true}, "请先登录")
			c.Abort()
			return
		}

		var DB_LinksToken DB.DB_LinksToken
		//这里如果报错  invalid memory address or nil pointer dereference   可能是配置文件数据库配置北山,global.GVA_DB 值为空
		err := global.GVA_DB.Model(DB.DB_LinksToken{}).Where("Token = ?", Token).First(&DB_LinksToken).Error
		// 没查到数据 或状态不正常
		if err != nil || DB_LinksToken.Status != 1 {
			response.FailTokenErr(c, gin.H{}, "令牌已失效")
			c.Abort()
			return
		}

		if DB_LinksToken.LoginAppid != constant.APPID_Web用户中心 {
			response.FailTokenErr(c, gin.H{}, "非WebUser令牌,请重新登录")
			c.Abort()
			return
		}
		db := *global.GVA_DB
		var 局_网页用户中心配置 dbm.DB_AppInfoWebUser
		局_网页用户中心配置, err = service.NewAppInfoWebUser(c, &db).Info(D到整数(DB_LinksToken.AppIdEx))
		if err != nil || 局_网页用户中心配置.Status != 1 {
			response.FailTokenErr(c, gin.H{"reload": true}, "应用未开放网页用户中心,请联系管理员")
			c.Abort()
			return
		}

		data, err := c.GetRawData() //GetRawData只能使用一次
		if err != nil {
			response.FailTokenErr(c, gin.H{}, "参数错误")
			c.Abort()
			return
		}
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data)) // 关键点 //通过这写回post数据,就可以多次读取了

		c.Set("DB_LinksToken", DB_LinksToken)
		c.Set("网页用户中心配置", 局_网页用户中心配置)
		//更新最后活动时间
		if time.Now().Unix()-DB_LinksToken.LastTime > 60 { //超过1分钟,更新最后活动时间
			global.GVA_DB.Model(DB.DB_LinksToken{}).Where("Id = ?", DB_LinksToken.Id).Updates(map[string]interface{}{"LastTime": int(time.Now().Unix()), "Ip": c.ClientIP()})
		}
		// 继续处理请求
		c.Next()
	}
}
