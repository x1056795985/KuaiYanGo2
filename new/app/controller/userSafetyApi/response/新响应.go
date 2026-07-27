package response

import (
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/encoding/gjson"
	"server/new/app/models/constant"
	"server/new/app/utils"
	"time"
)

func Ok(c *gin.Context) {
	X写入数据(c, 0, "ok", nil)
}
func OkMsg(c *gin.Context, msg string) {
	X写入数据(c, 0, msg, nil)
}
func OkData(c *gin.Context, Data interface{}) {
	X写入数据(c, 0, "ok", Data)
}

func Fail(c *gin.Context, status int) {
	X写入数据(c, status, "", nil)
}
func FailMsg(c *gin.Context, status int, msg string) {
	X写入数据(c, status, msg, nil)
}
func FailData(c *gin.Context, status int, msg string, Data interface{}) {
	X写入数据(c, status, msg, Data)
}

func X写入数据(c *gin.Context, 状态码 int, msg string, Data interface{}) {
	局_ctx := utils.Q取上下文(c)
	if 状态码 == 0 {
		状态码 = 局_ctx.C成功状态码
	}
	if msg == "" && constant.Status值键[状态码] != "" {
		msg = constant.Status值键[状态码]
	}
	局_ctx.X响应明文 = gjson.New("{}")
	局_ctx.X响应明文.Set("Time", time.Now().Unix())
	局_ctx.X响应明文.Set("Status", 状态码)
	局_ctx.X响应明文.Set("Msg", msg)
	if Data != nil {
		局_ctx.X响应明文.Set("Data", Data)
	}
}
