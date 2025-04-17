package Common

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"server/global"
	"server/structs/Http/response"
)

type Common struct {
}

// 统一反序列化参数
func (C *Common) ToJSON(c *gin.Context, obj any) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		// 获取validator.ValidationErrors类型的errors
		//20250411 发现检测有个问题 如果是逻辑型值为false的参数   参数开启了required 必填 他也会报错 参数不存在 解决办法,不校验required
		errs, ok := err.(validator.ValidationErrors)
		errStr := ""
		if !ok {
			errStr = "参数错误:" + err.Error() //	// 非validator.ValidationErrors类型错误直接返回
		} else {
			for _, v := range errs.Translate(global.Trans) { // validator.ValidationErrors类型错误则进行翻译
				errStr += v + ","
			}
		}
		response.FailWithMessage(errStr, c)
		return false
	}
	return true
}
