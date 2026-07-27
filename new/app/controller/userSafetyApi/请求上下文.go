package userSafetyApi

import (
	"github.com/gin-gonic/gin"
	"server/new/app/models/common"
	"server/new/app/utils"
)

// 请求上下文 类型别名,统一使用 common.Q请求_上下文
type 请求上下文 = common.Q请求_上下文

// 取上下文 从 gin.Context 取出请求上下文(统一使用 utils.Q取上下文)
func 取上下文(c *gin.Context) *请求上下文 {
	return utils.Q取上下文(c)
}
