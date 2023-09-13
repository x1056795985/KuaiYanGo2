// 返回json的结构体 声明
package response

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	DB "server/structs/db"
	"server/utils"
	"strings"
	"time"
)

// 回复json结构体
type 请求响应_加密包 struct {
	A密文 string `json:"a"`
	B签名 string `json:"b"`
}

func 响应加密处理(c *gin.Context, 明文Json string) {
	局_临时通用, ok := c.Get("AppInfo")
	if !ok {
		c.JSON(http.StatusOK, 明文Json)
		return
	}

	AppInfo := 局_临时通用.(DB.DB_AppInfo)
	if AppInfo.CryptoType <= 1 {
		c.JSON(http.StatusOK, 明文Json)
		return
	}

	AesKey := c.GetString("局_CryptoKeyAes")

	var 局_加密编码后 string
	var 签名结果 string

	if AppInfo.CryptoType == 2 {
		//方式2就直接AES加密就好了 然后AES签名
		临时 := utils.Aes加密_cbc192(明文Json, AesKey)
		局_加密编码后 = base64.StdEncoding.EncodeToString(临时)
		签名结果 = 签名_Aes(局_加密编码后, AesKey)
	} else if AppInfo.CryptoType == 3 {

		if c.GetBool("RSA强制") {
			//这种情况 随机AES密钥然后加密   签名结果就是Aes的密钥的RSA加密 因为Rsa无法加密长数据所以只加密密匙
			//AesKey = utils.W文本_取随机字符串(24)
			局_AES随机密匙字节集 := make([]byte, 24)
			_, _ = rand.Read(局_AES随机密匙字节集)
			临时 := utils.Aes加密_cbc192密匙字节数组(明文Json, 局_AES随机密匙字节集)
			局_加密编码后 = base64.StdEncoding.EncodeToString(临时)
			签名结果 = utils.RSA私钥加密([]byte(AppInfo.CryptoKeyPrivate), 局_AES随机密匙字节集) //RSa加密的是AES密钥
		} else { //不重要的信息,AES加密然后就还AES签名即可 提高效率
			临时 := utils.Aes加密_cbc192(明文Json, AesKey)
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
	return strings.ToUpper(utils.Md5String(base64后明文 + AesKey))
}

func X响应状态(c *gin.Context, 状态码 int) {
	j1 := 请求响应_X响应状态{time.Now().Unix(), 状态码, Status值键[状态码]}

	if c.GetString("局_CryptoKeyAes") == "" {
		//没有通讯key直接返回明文
		c.JSON(http.StatusOK, j1)
	} else {
		json明文, _ := json.Marshal(j1)
		响应加密处理(c, string(json明文))
	}
}

func X响应状态消息(c *gin.Context, 状态码 int, Msg string) {
	j1 := 请求响应_X响应状态{time.Now().Unix(), 状态码, Msg}

	if c.GetString("局_CryptoKeyAes") == "" {
		//没有通讯key直接返回明文
		c.JSON(http.StatusOK, j1)
	} else {
		json明文, _ := json.Marshal(j1)
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
	if c.GetString("局_CryptoKeyAes") == "" {
		//没有通讯key直接返回明文
		c.JSON(http.StatusOK, 局_明文结构)
	} else {
		json明文, _ := json.Marshal(局_明文结构)
		响应加密处理(c, string(json明文))
	}
}

// 回复json结构体
type 请求响应_X响应成功带数据 struct {
	Data interface{} `json:"Data"`
	请求响应_X响应状态
}
