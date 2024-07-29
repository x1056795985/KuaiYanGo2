// 返回json的结构体 声明
package response

import (
	"EFunc/utils"
	"encoding/base64"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/valyala/fastjson"
	"net/http"
	"server/Service/Ser_Js"
	DB "server/structs/db"
	. "server/utils"
	utils2 "server/utils"
	"strings"
	"time"
)

// 回复json结构体
type 请求响应_加密包 struct {
	A密文 string `json:"a"`
	B签名 string `json:"b"`
}

func 响应不加密处理(c *gin.Context, 明文Json string) {
	明文Json = apiHook之后(c, 明文Json)
	c.String(http.StatusOK, 明文Json)
}
func 响应加密处理(c *gin.Context, 明文Json string) {

	明文Json = apiHook之后(c, 明文Json)

	局_临时通用, ok := c.Get("AppInfo")
	if !ok {
		c.JSON(http.StatusOK, 明文Json)
		return
	}

	AppInfo := 局_临时通用.(DB.DB_AppInfo)
	if AppInfo.CryptoType <= 1 { //明文
		c.String(http.StatusOK, 明文Json)
		return
	}

	AesKey := c.GetString("局_CryptoKeyAes")

	var 局_加密编码后 string
	var 签名结果 string

	if AppInfo.CryptoType == 2 {
		//方式2就直接AES加密就好了 然后AES签名
		临时 := utils2.Aes加密_cbc192(明文Json, AesKey)
		局_加密编码后 = base64.StdEncoding.EncodeToString(临时)
		签名结果 = 签名_Aes(局_加密编码后, AesKey)
	} else if AppInfo.CryptoType == 3 {

		if c.GetBool("RSA强制") {
			//这种情况 随机AES密钥然后加密   签名结果就是Aes的密钥的RSA加密 因为Rsa无法加密长数据所以只加密密匙

			/*		局_AES随机密匙字节集 := make([]byte, 24)
					_, _ = rand.Read(局_AES随机密匙字节集)*/
			局_AES随机密匙字节集 := []byte(utils.W文本_取随机字符串(24)) // 因为js暂时无法公钥解密出字节数组的密钥,所以暂时改为文本字符串,方便更多语言对接

			临时 := utils2.Aes加密_cbc192密匙字节数组(明文Json, 局_AES随机密匙字节集)
			局_加密编码后 = base64.StdEncoding.EncodeToString(临时)
			签名结果 = utils2.RSA私钥加密([]byte(AppInfo.CryptoKeyPrivate), 局_AES随机密匙字节集) //RSa加密的是AES密钥
		} else { //不重要的信息,AES加密然后就还AES签名即可 提高效率
			临时 := utils2.Aes加密_cbc192(明文Json, AesKey)
			局_加密编码后 = base64.StdEncoding.EncodeToString(临时)
			签名结果 = 签名_Aes(局_加密编码后, AesKey)
		}

	}

	c.JSON(http.StatusOK, 请求响应_加密包{
		局_加密编码后,
		签名结果,
	})

}

func 签名_Aes(base64后明文 string, AesKey string) string {
	return strings.ToUpper(Md5String(base64后明文 + AesKey))
}

func X响应状态(c *gin.Context, 状态码 int) {
	局_明文结构 := 请求响应_X响应状态{time.Now().Unix(), 状态码, Status值键[状态码]}

	json明文, _ := json.Marshal(局_明文结构)
	if c.GetString("局_CryptoKeyAes") == "" {
		//没有通讯key直接返回明文
		响应不加密处理(c, string(json明文))
	} else {
		响应加密处理(c, string(json明文))
	}
}

func X响应状态消息(c *gin.Context, 状态码 int, Msg string) {
	局_明文结构 := 请求响应_X响应状态{time.Now().Unix(), 状态码, Msg}
	json明文, _ := json.Marshal(局_明文结构)
	if c.GetString("局_CryptoKeyAes") == "" {
		//没有通讯key直接返回明文
		响应不加密处理(c, string(json明文))
	} else {
		响应加密处理(c, string(json明文))
	}
}

// 回复json结构体
type 请求响应_X响应状态 struct {
	Time   int64  `json:"Time"`
	Status int    `json:"Status"`
	Msg    string `json:"Msg"`
}

func X响应状态带数据(c *gin.Context, 状态码 int, Data interface{}) {
	var 局_明文结构 = 请求响应_X响应成功带数据{}
	局_明文结构.请求响应_X响应状态 = 请求响应_X响应状态{time.Now().Unix(), 状态码, Status值键[状态码]}
	局_明文结构.Data = Data

	json明文, _ := json.Marshal(局_明文结构)
	if c.GetString("局_CryptoKeyAes") == "" {
		//没有通讯key直接返回明文
		响应不加密处理(c, string(json明文))
	} else {
		响应加密处理(c, string(json明文))
	}
}

// 回复json结构体
type 请求响应_X响应成功带数据 struct {
	Data interface{} `json:"Data"`
	请求响应_X响应状态
}

func apiHook之后(c *gin.Context, json明文 string) string {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	var Api = c.GetString("局_Api")
	局_临时通用, _ := c.Get("AppInfo")
	AppInfo = 局_临时通用.(DB.DB_AppInfo)
	局_临时通用, _ = c.Get("局_在线信息")
	if 局_临时通用 != nil { //接口 GetToken  没有在线信息这个值为空
		局_在线信息 = 局_临时通用.(DB.DB_LinksToken)
	}

	//==========================ApiHook之后====================================
	if utils.W文本_是否包含关键字(AppInfo.ApiHook, `"`+Api+`"`) { //先判断Api是否需要Hook
		//{"UserLogin":{"Before":"hook登录前","After":"hook登录后"}}
		JSON, err := fastjson.Parse(AppInfo.ApiHook)
		if err != nil {
			return json明文
		}
		局_hookAfter := ""
		局_hookAfter = string(JSON.GetStringBytes(Api, "After"))
		if len(局_hookAfter) == 0 {
			return json明文
		}

		json明文, err = Ser_Js.JS引擎初始化_ApiHook处理(&AppInfo, &局_在线信息, 局_hookAfter, json明文, c)
		if err != nil {
			局_明文结构 := 请求响应_X响应状态{time.Now().Unix(), Status_操作失败, err.Error()}
			json明文字节集, _ := json.Marshal(局_明文结构)
			return string(json明文字节集)

		}
	}
	return json明文
	//=============================ApiHook结束=================================
}
