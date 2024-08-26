// 返回加密结果
package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
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
		data,
		msg,
	})

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
