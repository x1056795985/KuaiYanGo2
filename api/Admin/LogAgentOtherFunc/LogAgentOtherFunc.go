package LogAgentOtherFunc

import (
	"github.com/gin-gonic/gin"
	"server/Service/Ser_User"
	"server/global"
	"server/new/app/logic/common/agent"
	"server/structs/Http/response"
	DB "server/structs/db"
	"strconv"
	"time"
)

type Api struct{}

type 结构请求_GetDB_getList struct {
	Page     int      `json:"Page"`     // 页
	Size     int      `json:"Size"`     // 页数量
	Type     int      `json:"Type"`     // 关键字类型
	Keywords string   `json:"Keywords"` // 关键字
	Order    int      `json:"Order"`    // 0 倒序 1 正序
	Time     []string `json:"Time"`     // 开始时间 结束时间
	Count    int64    `json:"Count"`    // 总数
	Func     int64    `json:"Func"`     //操作功能id
}

// GetDB_LogRegisterKaList

func (a *Api) GetLogList(c *gin.Context) {
	var 请求 结构请求_GetDB_getList
	//{"Type":2,"Size":10,"Page":1,"Status":1,"keywords":"1"}
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	局_DB := global.GVA_DB.Model(DB.DB_LogAgentOtherFunc{})

	if 请求.Func < 0 {
		局_DB.Where("Func = ? ", 请求.Func)
	}

	if 请求.Time != nil && len(请求.Time) == 2 && 请求.Time[0] != "" && 请求.Time[1] != "" {
		开始时间, _ := strconv.ParseInt(请求.Time[0], 10, 64)
		结束时间, _ := strconv.ParseInt(请求.Time[1], 10, 64)
		局_DB.Where("Time > ?", 开始时间).Where("Time < ?", 结束时间+86400)
	}

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //代理id
			局_代理Uid := Ser_User.User用户名取id(请求.Keywords)
			if 局_代理Uid == 0 {
				response.FailWithMessage("代理账号错误", c)
				return
			}
			局_DB.Where("AgentUid  = ? ", 局_代理Uid)
		case 2: //用户user

			局_DB.Where("AppUser like  ?", "%"+请求.Keywords+"%")
		case 3: //ip
			局_DB.Where("Ip like ? ", "%"+请求.Keywords+"%")
		case 4: //信息
			局_DB.Where("Note like ?", "%"+请求.Keywords+"%")
		}
	}
	if 请求.Order == 1 {
		局_DB.Order("Id ASC")
	} else {
		局_DB.Order("Id DESC")
	}
	var LogAgentOtherFunc []DB.DB_LogAgentOtherFunc
	var 总数 int64
	//Count(&总数) 必须放在where 后面 不然值会被清0
	if 请求.Count > 500000 { //如果数据大于50万 直接使用,不重新查询了 优化速度
		总数 = 请求.Count
	} else {
		局_DB.Count(&总数)
	}
	err = 局_DB.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&LogAgentOtherFunc).Error
	//fmt.Println("用户总数%d", 总数, DB_LinksToken)
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		return
	}

	局_AgentIds := make([]int, 0, len(LogAgentOtherFunc))
	for 索引, _ := range LogAgentOtherFunc {
		局_AgentIds = append(局_AgentIds, LogAgentOtherFunc[索引].AgentUid)
	}
	局_MapUId_User := Ser_User.Id取User_批量(局_AgentIds)
	局_Map代理ID_功能 := agent.L_agent.Q取全部代理功能ID_MAP(c)
	局_DB_LogAgentOtherFunc扩展 := make([]DB_LogAgentOtherFunc扩展, len(LogAgentOtherFunc))
	for 索引, _ := range LogAgentOtherFunc {
		局_DB_LogAgentOtherFunc扩展[索引] = DB_LogAgentOtherFunc扩展{
			LogAgentOtherFunc[索引],
			局_MapUId_User[LogAgentOtherFunc[索引].AgentUid],
			局_Map代理ID_功能[LogAgentOtherFunc[索引].Func]}
	}

	response.OkWithDetailed(结构响应_GetDB_DB_LogAgentOtherFuncList{局_DB_LogAgentOtherFunc扩展, 总数}, "获取成功", c)
	return
}

type 结构响应_GetDB_DB_LogAgentOtherFuncList struct {
	List  interface{} `json:"List"`  // 列表
	Count int64       `json:"Count"` // 总数
}
type DB_LogAgentOtherFunc扩展 struct {
	DB.DB_LogAgentOtherFunc
	AgentUser string `json:"AgentUser"` // 总数
	FuncTxt   string `json:"FuncTxt"`   // 中文名称
}

type 结构请求_批量Delete struct {
	Id       []int  `json:"Id"`       //用户id数组
	Type     int    `json:"Type"`     //  1删除用户数组 2删除指定关键字 3清空 4删除7天前 5删除30天前 6删除90天前
	Keywords string `json:"Keywords"` //
}

// Del批量删除
func (a *Api) Delete(c *gin.Context) {
	var 请求 结构请求_批量Delete
	err := c.ShouldBindJSON(&请求)
	// 解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	var 影响行数 int64
	var db = global.GVA_DB.Model(DB.DB_LogAgentOtherFunc{})

	if 请求.Type <= 0 || 请求.Type > 7 {
		response.FailWithMessage("Type错误", c)
		return
	}

	// 1删除日志id数组 2删除指定用户 3清空 4删除7天前 5删除30天前 6删除90天前  7 关键字
	switch 请求.Type {
	case 1:
		if 请求.Type == 1 && len(请求.Id) == 0 {
			response.FailWithMessage("Id数组没有要删除的ID", c)
			return
		}
		影响行数 = db.Where("Id IN ? ", 请求.Id).Delete(DB.DB_LogAgentOtherFunc{}).RowsAffected
	case 2:
		影响行数 = db.Where("AppUser = ? ", 请求.Keywords).Delete(DB.DB_LogAgentOtherFunc{}).RowsAffected
	case 3: //清空
		影响行数 = db.Where("1=1").Delete(DB.DB_LogAgentOtherFunc{}).RowsAffected
	case 4: //删7天前
		影响行数 = db.Where("Time <  ?", time.Now().Unix()-604800).Delete(DB.DB_LogAgentOtherFunc{}).RowsAffected
	case 5: //删除30天前
		影响行数 = db.Where("Time <  ?", time.Now().Unix()-2592000).Delete(DB.DB_LogAgentOtherFunc{}).RowsAffected
	case 6: //删除90天前
		影响行数 = db.Where("Time <  ?", time.Now().Unix()-7776000).Delete(DB.DB_LogAgentOtherFunc{}).RowsAffected
	case 7: //删除关键字
		if len(请求.Keywords) == 0 {
			response.FailWithMessage("关键字不能为空", c)
			return
		}
		影响行数 = db.Where("Note LIKE ? ", "%"+请求.Keywords+"%").Delete(请求.Id).RowsAffected
	}

	if db.Error != nil {
		response.FailWithMessage("删除失败", c)
		return
	}

	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
	return
}
