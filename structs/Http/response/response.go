// 返回加密结果
package response

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"unicode"
)

// 回复json结构体
type Response struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

// 常量 回复状态码
const (
	ERROR       = 200
	SUCCESS     = 10000
	login登录状态失效 = 202
)

func Result(code int, data interface{}, msg string, c *gin.Context) {
	// 开始时间
	c.JSON(http.StatusOK, Response{
		code,
		代理端数据转小驼峰(c, data),
		msg,
	})

}

func 代理端数据转小驼峰(c *gin.Context, data interface{}) interface{} {
	if c == nil || c.Request == nil {
		return data
	}
	局_isAgentResponseCamel, ok := c.Get("isAgentResponseCamel")
	if !ok {
		return data
	}
	if 局_是否启用, ok := 局_isAgentResponseCamel.(bool); !ok || !局_是否启用 {
		return data
	}

	局_字节集, err := json.Marshal(data)
	if err != nil {
		return data
	}

	var 局_对象 interface{}
	if err = json.Unmarshal(局_字节集, &局_对象); err != nil {
		return data
	}
	return 递归转小驼峰(局_对象)
}

func 递归转小驼峰(data interface{}) interface{} {
	switch 局_值 := data.(type) {
	case map[string]interface{}:
		局_返回 := make(map[string]interface{}, len(局_值))
		for 局_键名, 局_成员 := range 局_值 {
			局_返回[文本转小驼峰(局_键名)] = 递归转小驼峰(局_成员)
		}
		return 局_返回
	case []interface{}:
		for 局_索引 := range 局_值 {
			局_值[局_索引] = 递归转小驼峰(局_值[局_索引])
		}
		return 局_值
	default:
		return data
	}
}

func 文本转小驼峰(key string) string {
	if key == "" {
		return key
	}
	switch strings.ToLower(key) {
	case "rmb":
		return "rmb"
	case "qq":
		return "qq"
	case "id":
		return "id"
	case "ids":
		return "ids"
	}

	局_文本 := key
	局_文本 = strings.ReplaceAll(局_文本, "Appid", "AppId")
	局_文本 = strings.ReplaceAll(局_文本, "ID", "Id")
	if strings.EqualFold(局_文本, "RMb") || strings.EqualFold(局_文本, "Rmb") {
		return "rmb"
	}

	局_rune数组 := []rune(局_文本)
	局_前缀长度 := 0
	for 局_前缀长度 < len(局_rune数组) && unicode.IsUpper(局_rune数组[局_前缀长度]) {
		局_前缀长度++
	}
	if 局_前缀长度 == 0 {
		return key
	}
	if 局_前缀长度 == 1 {
		局_rune数组[0] = unicode.ToLower(局_rune数组[0])
		return string(局_rune数组)
	}
	if 局_前缀长度 == len(局_rune数组) {
		for 局_索引 := range 局_rune数组 {
			局_rune数组[局_索引] = unicode.ToLower(局_rune数组[局_索引])
		}
		return string(局_rune数组)
	}
	for 局_索引 := 0; 局_索引 < 局_前缀长度-1; 局_索引++ {
		局_rune数组[局_索引] = unicode.ToLower(局_rune数组[局_索引])
	}
	return string(局_rune数组)
}

// 回复 操作成功
func Ok(c *gin.Context) {
	Result(SUCCESS, map[string]interface{}{}, "操作成功", c)
}

func OkWithMessage(message string, c *gin.Context) {
	Result(SUCCESS, map[string]interface{}{}, message, c)
}

func OkWithData(data interface{}, c *gin.Context) {
	Result(SUCCESS, data, "ok", c)
}

// 回复 成功 data  msa 信息
func OkWithDetailed(data interface{}, message string, c *gin.Context) {
	Result(SUCCESS, data, message, c)
}

// 回复 操作失败
func Fail(c *gin.Context) {
	Result(ERROR, map[string]interface{}{}, "操作失败", c)
}

// 回复操作失败 自定义消息
func FailWithMessage(message string, c *gin.Context) {
	Result(ERROR, map[string]interface{}{}, message, c)
}

// 回复操作失败 data  消息
func FailWithDetailed(data interface{}, message string, c *gin.Context) {
	Result(ERROR, data, message, c)
}

// 回复token错误 data  消息
func FailTokenErr(data interface{}, message string, c *gin.Context) {
	Result(login登录状态失效, data, message, c)
}
