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

func Result(c *gin.Context, code int, data interface{}, msg string) {
	// 开始时间
	c.JSON(http.StatusOK, Response{
		code,
		data,
		msg,
	})

}

// 回复 操作成功
func Ok(c *gin.Context) {
	Result(c, SUCCESS, map[string]interface{}{}, "操作成功")
}

func OkWithMessage(c *gin.Context, message string) {
	Result(c, SUCCESS, map[string]interface{}{}, message)
}

func OkWithData(c *gin.Context, data interface{}) {
	Result(c, SUCCESS, data, "ok")
}

// 回复 成功 data  msa 信息
func OkWithDetailed(c *gin.Context, data interface{}, message string) {
	Result(c, SUCCESS, data, message)
}

// 回复 操作失败
func Fail(c *gin.Context) {
	Result(c, ERROR, map[string]interface{}{}, "操作失败")
}

// 回复操作失败 自定义消息
func FailWithMessage(c *gin.Context, message string) {
	Result(c, ERROR, map[string]interface{}{}, message)
}

// 回复操作失败 自定义代码
func FailWithCode(c *gin.Context, Code int, message string) {
	Result(c, Code, map[string]interface{}{}, message)
}

// 回复操作失败 data  消息
func FailWithDetailed(c *gin.Context, data interface{}, message string) {
	Result(c, ERROR, data, message)
}

// 回复token错误 data  消息
func FailTokenErr(c *gin.Context, data interface{}, message string) {
	Result(c, login登录状态失效, data, message)
}
