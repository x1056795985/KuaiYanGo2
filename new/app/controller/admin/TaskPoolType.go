package controller

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common"

	"server/new/app/service"
	"server/structs/Http/response"
	"strconv"
)

type TaskPoolType struct {
	Common.Common
}

func NewTaskPoolTypeController() *TaskPoolType {
	return &TaskPoolType{}
}

// 修改app排序
func (C *TaskPoolType) SetSort(c *gin.Context) {
	var 请求 struct {
		Id   int   `json:"Id"`
		Sort int64 `json:"Sort"`
	}
	//解析失败
	if !C.ToJSON(c, &请求) {
		return
	}
	tx := *global.GVA_DB
	var S = service.NewTaskPoolType(c, &tx)

	row, err := S.Update(请求.Id, map[string]interface{}{"Sort": 请求.Sort})
	if err != nil {
		response.FailWithMessage(err.Error(), c)
	}
	response.OkWithMessage("操作成功,数量:"+strconv.Itoa(int(row)), c)
	return
}
