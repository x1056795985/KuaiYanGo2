package middleware

import (
	. "EFunc/utils"
	"encoding/base64"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/valyala/fastjson"
	"net/http"
	"server/Service/Ser_Js"
	"server/new/app/models/common"
	"server/new/app/models/constant"
	"server/new/app/utils"
	utils2 "server/utils"
	"strings"
	"time"
)

// 请求响应_加密包 回复加密json结构体
type 请求响应_加密包 struct {
	A密文 string `json:"a"`
	B签名 string `json:"b"`
}

// 回复json结构体
type 请求响应_X响应状态 struct {
	Time   int64  `json:"Time"`
	Status int    `json:"Status"`
	Msg    string `json:"Msg"`
}

// C处理响应数据 出口中间件
// handler 通过 response 包把明文写入 ctx.X响应明文。
// 此中间件在请求链结束后执行 apiHook之后脚本并尝试加密，加密失败则返回明文。
func C处理响应数据() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		ctx := utils.Q取上下文(c)
		if ctx == nil || ctx.X响应明文 == nil {
			return
		}

		明文Json := ctx.X响应明文.String()
		if 明文Json == "" {
			return
		}

		明文Json = apiHook之后(c, ctx, 明文Json)

		if 加密包, ok := J尝试加密响应(ctx, 明文Json); ok {
			c.JSON(http.StatusOK, 加密包)
			return
		}

		c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(明文Json))
	}
}

// J尝试加密响应 将所有加密失败统一转换为 ok=false，由调用方回退到明文响应。
func J尝试加密响应(ctx *common.Q请求_上下文, 明文Json string) (加密包 请求响应_加密包, ok bool) {
	defer func() {
		if recover() != nil {
			加密包 = 请求响应_加密包{}
			ok = false
		}
	}()

	AppInfo := ctx.AppInfo
	if AppInfo.CryptoType <= 1 {
		return 加密包, false
	}

	if AppInfo.CryptoType == 3 && ctx.RSA强制 {
		局_AES随机密匙字节集 := []byte(W文本_取随机字符串(24))
		密文 := utils2.Aes加密_cbc192密匙字节数组(明文Json, 局_AES随机密匙字节集)
		加密包.A密文 = base64.StdEncoding.EncodeToString(密文)
		加密包.B签名 = utils2.RSA私钥加密([]byte(AppInfo.CryptoKeyPrivate), 局_AES随机密匙字节集)
		return 加密包, 加密包.A密文 != "" && 加密包.B签名 != ""
	}

	if (AppInfo.CryptoType != 2 && AppInfo.CryptoType != 3) || len(ctx.CryptoKeyAes) != 24 {
		return 加密包, false
	}
	密文 := utils2.Aes加密_cbc192(明文Json, ctx.CryptoKeyAes)
	加密包.A密文 = base64.StdEncoding.EncodeToString(密文)
	加密包.B签名 = 签名_Aes(加密包.A密文, ctx.CryptoKeyAes)
	return 加密包, 加密包.A密文 != "" && 加密包.B签名 != ""
}

func 签名_Aes(base64后明文 string, AesKey string) string {
	return strings.ToUpper(utils2.Md5String(base64后明文 + AesKey))
}

// apiHook之后 在响应明文返回前执行ApiHook之后脚本
func apiHook之后(c *gin.Context, ctx *common.Q请求_上下文, json明文 string) string {
	AppInfo := ctx.AppInfo
	Api := ctx.Api

	if !W文本_是否包含关键字(AppInfo.ApiHook, `"`+Api+`"`) {
		return json明文
	}

	//{"UserLogin":{"Before":"hook登录前","After":"hook登录后"}}
	//用 fastjson 解析 ApiHook 配置,取出 After 脚本名
	JSON, err := fastjson.Parse(AppInfo.ApiHook)
	if err != nil {
		return json明文
	}
	局_hookAfter := string(JSON.GetStringBytes(Api, "After"))
	if 局_hookAfter == "" {
		return json明文
	}

	局_在线信息 := ctx.Z在线信息
	局_明文, err := Ser_Js.JS引擎初始化_ApiHook处理(&AppInfo, &局_在线信息, 局_hookAfter, json明文, c)
	if err != nil {
		局_明文结构 := 请求响应_X响应状态{time.Now().Unix(), constant.Status_操作失败, err.Error()}
		json明文字节集, _ := json.Marshal(局_明文结构)
		return string(json明文字节集)
	}
	return 局_明文
}
