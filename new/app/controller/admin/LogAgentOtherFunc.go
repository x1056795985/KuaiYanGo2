package controller

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/logic/common/agent"
	"server/new/app/models/db"
	"server/new/app/models/request"
	. "server/new/app/models/response"
	"server/new/app/service"
	"server/structs/Http/response"
	"strconv"
)

// LogAgentOtherFunc 代理操作日志
type LogAgentOtherFunc struct {
	Common.Common
}

func NewLogAgentOtherFuncController() *LogAgentOtherFunc {
	return &LogAgentOtherFunc{}
}

// Info 获取详情
func (C *LogAgentOtherFunc) Info(c *gin.Context) {
	var 请求 request.Id
	if !C.ToJSON(c, &请求) {
		return
	}

	var S = service.S_LogAgentOtherFunc{}
	tx := *global.GVA_DB
	info, err := S.Info(&tx, 请求.Id)
	if err != nil {
		response.FailWithMessage("获取失败,可能不存在", c)
		return
	}
	response.OkWithDetailed(info, "获取成功", c)
}

type 请求_LogAgentOtherFuncGetList struct {
	request.List
	Func int64 `json:"func"` // 操作功能id
}

// GetList 代理操作日志列表
func (C *LogAgentOtherFunc) GetList(c *gin.Context) {
	var 请求 请求_LogAgentOtherFuncGetList
	if !C.ToJSON(c, &请求) {
		return
	}

	var S = service.S_LogAgentOtherFunc{}
	tx := *global.GVA_DB
	总数, dataList, err := S.GetList(&tx, service.LogAgentOtherFuncListRequest{
		List: 请求.List,
		Func: 请求.Func,
	})
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		return
	}

	局_AgentIds := make([]int, 0, len(dataList))
	for 索引, _ := range dataList {
		局_AgentIds = append(局_AgentIds, dataList[索引].AgentUid)
	}
	type DB_LogAgentOtherFunc扩展 struct {
		db.DB_LogAgentOtherFunc
		AgentUser string `json:"AgentUser"` // 总数
		FuncTxt   string `json:"FuncTxt"`   // 中文名称
	}
	局_MapUId_User, _ := service.NewUser(c, &tx).Infos(map[string]interface{}{"Id": 局_AgentIds})

	局_Map代理ID_功能 := agent.L_agent.Q取全部代理功能ID_MAP(c)
	局_DB_LogAgentOtherFunc扩展 := make([]DB_LogAgentOtherFunc扩展, len(dataList))
	for 索引, _ := range dataList {
		局_DB_LogAgentOtherFunc扩展[索引] = DB_LogAgentOtherFunc扩展{
			dataList[索引],
			"",
			局_Map代理ID_功能[dataList[索引].Func]}

		for _, 用户信息 := range 局_MapUId_User {
			if 用户信息.Id == dataList[索引].AgentUid {
				局_DB_LogAgentOtherFunc扩展[索引].AgentUser = 用户信息.User
			}
		}
	}

	response.OkWithDetailed(GetList{List: 局_DB_LogAgentOtherFunc扩展, Count: 总数}, "获取成功", c)
}

type 请求_LogAgentOtherFuncBatchDelete struct {
	Id       []int  `json:"id"`
	Type     int    `json:"type"`
	Keywords string `json:"keywords"`
}

// Delete 批量删除
func (C *LogAgentOtherFunc) Delete(c *gin.Context) {
	var 请求 请求_LogAgentOtherFuncBatchDelete
	if !C.ToJSON(c, &请求) {
		return
	}

	var S = service.S_LogAgentOtherFunc{}
	tx := *global.GVA_DB
	影响行数, err := S.BatchDelete(&tx, 请求.Id, 请求.Type, 请求.Keywords)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
}
