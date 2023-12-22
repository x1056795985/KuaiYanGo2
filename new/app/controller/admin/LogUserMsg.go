package controller

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/service"
	"server/structs/Http/response"
)

type LogUserMsg struct {
}

func NewLogUserMsgController() *LogUserMsg {
	return &LogUserMsg{}
}

func (s *LogUserMsg) S删除重复消息(c *gin.Context) {
	var S = service.S_LogUserMsg{}
	tx := *global.GVA_DB
	err := S.S删除重复消息(&tx)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
	}
	response.Ok(c)

}
