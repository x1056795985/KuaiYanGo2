package middleware

import (
	. "EFunc/utils"
	"bytes"
	"encoding/base64"
	json2 "encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/encoding/gjson"
	"io"
	"server/app/controller/userSafetyApi/response"
	"server/app/global"
	"server/app/logic/common/jsEngine"
	"server/app/logic/common/log"
	"server/app/models/common"
	"server/app/models/constant"
	dbm "server/app/models/db"
	"server/app/service"
	"server/app/utils"
	utils2 "server/app/utils"
	"strings"
	"time"
)

// 路由名称解密函数指针,由 router/userSafetyApi 包在初始化时注入
var F解密Api名称 func(c *gin.Context, Api string) (string, bool)

// Token有效的才放行,否则返回Token失效
func UserApi解密() gin.HandlerFunc {
	return func(c *gin.Context) {

		Token := c.Request.Header.Get("Token")
		if Token == "" {
			c.Next()
			return
		}
		var err error
		局_ctx := utils.Q取上下文(c)
		AppInfo := 局_ctx.AppInfo
		db := *global.GVA_DB
		//先检查在线信息
		局_ctx.Z在线信息, err = service.NewLinksToken(c, &db).InfoToken(Token)
		if err != nil || 局_ctx.Z在线信息.LoginAppid != AppInfo.AppId {
			response.Fail(c, constant.Status_Token无效)
			c.Abort()
			return
		}
		if 局_ctx.Z在线信息.Status != 1 {
			response.FailData(c, constant.Status_Token已注销, "", gin.H{"LogoutCode": 局_ctx.Z在线信息.LogoutCode})
			c.Abort()
			return
		}
		局_map := make(map[string]interface{}, 2)

		if time.Now().Unix()-局_ctx.Z在线信息.LastTime > 60 {
			局_map["LastTime"] = time.Now().Unix()
			局_ctx.Z在线信息.LastTime = time.Now().Unix()
		}
		if 局_ctx.Z在线信息.Ip != c.ClientIP() {
			局_map["Ip"] = c.ClientIP()
			局_ctx.Z在线信息.Ip = c.ClientIP()
		}
		if len(局_map) > 0 {
			go service.NewLinksToken(c, &db).Update(局_ctx.Z在线信息.Id, 局_map)
		}
		// 在线信息获取完毕,解密
		局_临时字节集, err := c.GetRawData()
		if err != nil {
			response.Fail(c, constant.Status_参数错误)
			c.Abort()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(局_临时字节集))
		//先处理加密包
		var 结构加密包 struct {
			A密文 string `json:"a"`
			B签名 string `json:"b"`
		}
		if 局_ctx.AppInfo.CryptoType == 3 || 局_ctx.AppInfo.CryptoType == 2 {
			if err = json2.Unmarshal(局_临时字节集, &结构加密包); err != nil {
				response.Fail(c, constant.Status_参数错误)
				c.Abort()
				return
			}
			局_ctx.B值长度 = len(结构加密包.B签名)
		}

		局_明文 := ""

		if 局_ctx.AppInfo.CryptoType == 3 && len(结构加密包.B签名) > 32 {
			//rsa混合模式,使用appRsa私钥解密
			局_签名解密, _ := base64.StdEncoding.DecodeString(结构加密包.B签名)
			局_临时AES密匙 := utils2.Rsa私钥解密2([]byte(局_ctx.AppInfo.CryptoKeyPrivate), 局_签名解密)
			局_明文 = utils2.Aes解密_cbc192字节集(B编码_BASE64解码(结构加密包.A密文), 局_临时AES密匙)
		} else if 局_ctx.AppInfo.CryptoType == 3 && len(结构加密包.B签名) == 32 {
			//rsa混合模式,使用aes 使用在线信息的aes密钥
			局_明文 = utils2.Aes解密_cbc192(B编码_BASE64解码(结构加密包.A密文), 局_ctx.Z在线信息.CryptoKeyAes)
		} else if 局_ctx.AppInfo.CryptoType == 2 {
			//aes模式 使用应用信息的aes密钥
			局_明文 = utils2.Aes解密_cbc192(B编码_BASE64解码(结构加密包.A密文), 局_ctx.AppInfo.CryptoKeyAes)
		} else if 局_ctx.AppInfo.CryptoType == 1 {
			//不加密
			局_明文 = string(局_临时字节集)
		}
		if len(局_明文) == 0 {
			response.Fail(c, constant.Status_加解密失败)
			c.Abort()
			return
		}
		//开始安全性检查,校验签名
		if len(结构加密包.B签名) == 32 {
			期望签名 := strings.ToUpper(utils2.Md5String(结构加密包.A密文 + S三元(局_ctx.AppInfo.CryptoType == 3, 局_ctx.Z在线信息.CryptoKeyAes, 局_ctx.AppInfo.CryptoKeyAes)))
			if strings.ToUpper(结构加密包.B签名) != 期望签名 {
				go log.L_log.S写风控日志(c, 局_ctx.Z在线信息.Id, log.Log风控类型_Api异常调用, 局_ctx.Z在线信息.User, c.ClientIP(), "用户发送错误签名封包,可能在尝试破解")
				response.Fail(c, constant.Status_签名错误)
				c.Abort()
				return
			}
		}
		if 局_明文 == "" {
			response.Fail(c, constant.Status_加解密失败)
			c.Abort()
			return
		}
		局_ctx.Q请求明文 = gjson.New(局_明文)
		if 局_ctx.Q请求明文.IsNil() {
			response.Fail(c, constant.Status_加解密失败)
			c.Abort()
			return
		}
		if 局_int := 校验请求基础字段(c, 局_ctx.AppInfo, 局_ctx.Q请求明文); 局_int > 0 {
			response.Fail(c, 局_int)
			c.Abort()
			return
		}
		局_ctx.W无Token请求 = false
		局_ctx.CryptoKeyAes = 局_ctx.Z在线信息.CryptoKeyAes
		局_ctx.Api = 局_ctx.Q请求明文.Get("Api").String()
		局_ctx.C成功状态码 = 局_ctx.Q请求明文.Get("Status").Int()
		c.Next()
	}
}

func UserApi无Token解密() gin.HandlerFunc {
	return func(c *gin.Context) {

		if c.Request.Header.Get("Token") != "" {
			c.Next()
			return
		}
		局_ctx := utils.Q取上下文(c)
		局_临时字节集, _ := c.GetRawData()
		c.Request.Body = io.NopCloser(bytes.NewBuffer(局_临时字节集))
		//先处理加密包
		var 结构加密包 struct {
			A密文 string `json:"a"`
			B签名 string `json:"b"`
		}

		if 局_ctx.AppInfo.CryptoType == 3 || 局_ctx.AppInfo.CryptoType == 2 {
			if err := json2.Unmarshal(局_临时字节集, &结构加密包); err != nil {
				response.Fail(c, constant.Status_参数错误)
				c.Abort()
				return
			}
			局_ctx.B值长度 = len(结构加密包.B签名)
		}

		局_明文 := ""
		if 局_ctx.AppInfo.CryptoType == 3 {
			局_签名解密, _ := base64.StdEncoding.DecodeString(结构加密包.B签名)
			局_临时AES密匙 := utils2.Rsa私钥解密2([]byte(局_ctx.AppInfo.CryptoKeyPrivate), 局_签名解密)
			局_明文 = utils2.Aes解密_cbc192字节集(B编码_BASE64解码(结构加密包.A密文), 局_临时AES密匙)
		} else if 局_ctx.AppInfo.CryptoType == 2 {
			局_明文 = utils2.Aes解密_cbc192(B编码_BASE64解码(结构加密包.A密文), 局_ctx.AppInfo.CryptoKeyAes)
		} else if 局_ctx.AppInfo.CryptoType == 1 {
			局_明文 = string(局_临时字节集)
		}
		if 局_明文 == "" {
			response.Fail(c, constant.Status_加解密失败)
			c.Abort()
			return
		}
		局_ctx.Q请求明文 = gjson.New(局_明文)
		if 局_ctx.Q请求明文.IsNil() {
			response.Fail(c, constant.Status_加解密失败)
			c.Abort()
			return
		}

		if 局_int := 校验请求基础字段(c, 局_ctx.AppInfo, 局_ctx.Q请求明文); 局_int > 0 {
			response.Fail(c, 局_int)
			c.Abort()
			return
		}
		局_ctx.RSA强制 = true
		局_ctx.W无Token请求 = true
		局_ctx.Api = 局_ctx.Q请求明文.Get("Api").String()
		局_ctx.C成功状态码 = 局_ctx.Q请求明文.Get("Status").Int()
		c.Next()

	}
}

var 集_UserAPi路由强制RSA = map[string]int{
	"GetToken":            1,
	"UserLogin":           1,
	"UserReduceMoney":     1,
	"UserReduceVipNumber": 1,
	"UserReduceVipTime":   1,
	"GetVipData":          1,
}

func J解密Api名称() gin.HandlerFunc {
	return func(c *gin.Context) {
		局_ctx := utils.Q取上下文(c)
		局_ctx.Api = strings.TrimSpace(局_ctx.Api)
		局_Api, ok := F解密Api名称(c, 局_ctx.Api)
		if ok {
			局_ctx.Api = 局_Api
		}

		// 判断api是否强制rsa
		局_int, ok := 集_UserAPi路由强制RSA[局_ctx.Api]
		局_ctx.RSA强制 = ok && 局_int == 1

		// 强制RSA加密的Api,不允许使用AES加密方式
		if 局_ctx.AppInfo.CryptoType == 3 && 局_ctx.RSA强制 && 局_ctx.B值长度 == 32 { //强制RSA,但是签名长度不对,直接返回错误
			response.FailMsg(c, constant.Status_加解密失败, "该接口必须使用rsa加密")
			c.Abort()
			return
		}

		// apiHook之前: 在Api名称解密后、handler执行前执行ApiHook之前脚本
		if err := apiHook之前(c, 局_ctx); err != nil {
			return
		}

		c.Next()
	}
}

// apiHook之前 在handler执行前运行ApiHook之前脚本,修改请求明文
func apiHook之前(c *gin.Context, ctx *common.Q请求_上下文) error {
	AppInfo := ctx.AppInfo
	if !W文本_是否包含关键字(AppInfo.ApiHook, `"`+ctx.Api+`"`) {
		return nil
	}
	局_hookBefore := W文本_取出中间文本(AppInfo.ApiHook, `"`+ctx.Api+`":{"Before":"`, `"`)
	if 局_hookBefore == "" {
		return nil
	}

	局_在线信息 := ctx.Z在线信息
	局_json明文, err := jsEngine.J脚本引擎_处理ApiHook(&AppInfo, &局_在线信息, 局_hookBefore, ctx.Q请求明文.String(), c)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, err.Error())
		c.Abort()
		return err
	}
	// 同步更新 gjson 对象
	ctx.Q请求明文 = gjson.New(局_json明文)
	return nil
}

func 校验请求基础字段(c *gin.Context, AppInfo dbm.DB_AppInfo, 请求 *gjson.Json) int {
	局_Time := 请求.Get("Time").Int()
	if AppInfo.PackTimeOut != 0 && int(time.Now().Unix())-局_Time > AppInfo.PackTimeOut {
		return constant.Status_封包超时
	}
	if 请求.Get("Status").Int() < 10000 {
		return constant.Status_状态码错误
	}
	return 0
}
