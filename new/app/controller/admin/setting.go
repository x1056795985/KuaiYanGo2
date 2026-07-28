package controller

import (
	"github.com/gin-gonic/gin"
	"server/new/app/global"
	"server/new/app/models/old/response"
	"server/new/app/service"
)

type Setting struct {
}

func NewSettingController() *Setting {
	return &Setting{}
}

func (s *Setting) Info(c *gin.Context) {
	var S = service.S_Setting{}

	data, err := S.Info(global.GVA_DB, "aaa")
	if err == nil {
		response.FailWithMessage(err.Error(), c)
	}
	response.OkWithDetailed(data, "操作成功", c)

}
