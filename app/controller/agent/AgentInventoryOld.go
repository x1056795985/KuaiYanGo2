package controller

import (
	"github.com/gin-gonic/gin"
	"server/app/global"
	"server/app/logic/common/agent"
	"server/app/logic/common/agentLevel"
	"server/app/logic/common/ka"
	"server/app/logic/common/log"
	dbm "server/app/models/db"
	"server/app/models/old/response"
	"server/app/service"
	"strconv"
	"time"
)

type AgentInventoryOld struct{}

func NewAgentInventoryController() *AgentInventoryOld {
	return &AgentInventoryOld{}
}

type Agent库存单Id请求 struct {
	Id int `json:"id"`
}

type Agent库存列表请求 struct {
	Page     int    `json:"page"`
	Size     int    `json:"size"`
	Status   int    `json:"status"`
	Type     int    `json:"type"`
	Keywords string `json:"keywords"`
	Order    int    `json:"order"`
}

type Agent库存列表项 struct {
	dbm.Db_Agent_库存卡包
	User        string `json:"user" gorm:"column:User;index;comment:用户名"`
	KaClassName string `json:"kaClassName" gorm:"column:KaClassName;index;comment:卡类名称"`
	AppId       int    `json:"appId" gorm:"column:AppId;应用Id"`
	AppName     string `json:"appName" gorm:"column:AppName;应用名称"`
}

type Agent库存撤回请求 struct {
	Id   int    `json:"id"`
	Num  int    `json:"num"`
	Note string `json:"note"`
}

type Agent库存发送请求 struct {
	SourceID int    `json:"sourceId"`
	Num      int    `json:"num"`
	ToUserId int    `json:"toUserId"`
	Note     string `json:"note"`
}

type Agent库存卡类树响应 struct {
	KaClassTree []ka.K可制卡类授权树形框结构 `json:"kaClassTree"`
}

type Agent可发送下级 struct {
	Id       int    `json:"id"`
	User     string `json:"user"`
	Disabled bool   `json:"disabled"`
}

func (A *AgentInventoryOld) GetAgentInventoryInfo(c *gin.Context) {
	var 请求 Agent库存单Id请求
	if err := c.ShouldBindJSON(&请求); err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}
	if service.NewAgentInventory(c, global.GVA_DB).Id取归属Uid(请求.Id) != c.GetInt("Uid") {
		response.FailWithMessage("只能查看自己的库存详细信息", c)
		return
	}

	var 局_卡包 dbm.Db_Agent_库存卡包
	if err := global.GVA_DB.Model(dbm.Db_Agent_库存卡包{}).Where("id = ?", 请求.Id).First(&局_卡包).Error; err != nil {
		response.FailWithMessage("查询详细信息失败", c)
		return
	}
	response.OkWithDetailed(局_卡包, "获取成功", c)
}

func (A *AgentInventoryOld) GetAgentInventoryList(c *gin.Context) {
	var 请求 Agent库存列表请求
	if err := c.ShouldBindJSON(&请求); err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	局_DB := global.GVA_DB.Model(dbm.Db_Agent_库存卡包{}).
		Where("(Uid = ? OR RegisterUserId = ? OR SourceUid=? )", c.GetInt("Uid"), c.GetInt("Uid"), c.GetInt("Uid"))
	局_DB2 := global.GVA_DB.Model(dbm.Db_Agent_库存卡包{}).
		Where("(Uid = ? OR RegisterUserId = ? OR SourceUid=? )", c.GetInt("Uid"), c.GetInt("Uid"), c.GetInt("Uid"))

	if 请求.Order == 1 {
		局_DB.Order("Id ASC")
		局_DB2.Order("Id ASC")
	} else {
		局_DB.Order("Id DESC")
		局_DB2.Order("Id DESC")
	}

	if 请求.Status == 1 {
		局_DB.Where("EndTime > ? and NumMax>Num", time.Now().Unix())
		局_DB2.Where("EndTime > ? and ai.NumMax>ai.Num", time.Now().Unix())
	} else if 请求.Status == 2 {
		局_DB.Where("EndTime < ?", time.Now().Unix())
		局_DB2.Where("EndTime < ?", time.Now().Unix())
	} else if 请求.Status == 3 {
		局_DB.Where("Num>=NumMax")
		局_DB2.Where("ai.Num>=ai.NumMax")
	}

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1:
			局_DB.Where("Id = ?", 请求.Keywords)
			局_DB2.Where("Id = ?", 请求.Keywords)
		case 2:
			局_Id := service.NewUser(c, global.GVA_DB).User用户名取id(请求.Keywords)
			局_DB.Where("Uid= ? ", 局_Id)
			局_DB2.Where("Uid= ? ", 局_Id)
		case 3:
			局_剩余数量, err := strconv.Atoi(请求.Keywords)
			if err != nil {
				response.FailWithMessage("库存剩余数量只能为整数", c)
				return
			}
			局_DB.Where("(NumMax-Num) > ?", 局_剩余数量)
			局_DB2.Where("(ai.NumMax-ai.Num) > ?", 局_剩余数量)
		case 4:
			局_DB.Where("LOCATE(?, Note)>0 ", 请求.Keywords)
			局_DB2.Where("LOCATE(?, ai.Note)>0 ", 请求.Keywords)
		}
	}

	var 局_总数 int64
	var 局_列表 []Agent库存列表项
	局_DB.Count(&局_总数)
	err := 局_DB2.Table("db_Agent_Inventory ai").
		Select("ai.*, u.User, kc.Name AS KaClassName,kc.AppId").
		Joins("LEFT JOIN db_User u ON ai.Uid = u.Id").
		Joins("LEFT JOIN db_Ka_Class kc ON ai.KaClassId = kc.Id").
		Omit("AppName").
		Limit(请求.Size).
		Offset((请求.Page - 1) * 请求.Size).
		Find(&局_列表).Error
	if err != nil {
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}

	局_AppMap := service.NewAppInfo(c, global.GVA_DB).AppInfo取map列表Int(true)
	for 局_索引 := range 局_列表 {
		局_列表[局_索引].AppName = 局_AppMap[局_列表[局_索引].AppId]
	}

	response.OkWithDetailed(Agent列表响应{List: 局_列表, Count: 局_总数}, "获取成功", c)
}

func (A *AgentInventoryOld) New库存购买(c *gin.Context) {
	var 请求 dbm.Db_Agent_库存卡包
	if err := c.ShouldBindJSON(&请求); err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	请求.Id = 0
	请求.Uid = c.GetInt("Uid")
	局_新卡包, err := agent.L_agent.New代理购买(c, 请求.Uid, 请求.KaClassId, 请求.NumMax, 请求.EndTime, 请求.Note, c.ClientIP())
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithMessage("操作成功", c)
	局_角色 := agentLevel.L_agentLevel.Q取Id代理级别(c, c.GetInt("Uid"))
	if 局_角色 == 0 {
		局_角色 = 4
	}
	局_创建用户名 := ""
	if 请求.Uid < 0 {
		局_创建用户名 = service.NewAdmin(c, global.GVA_DB).Id取User(请求.Uid)
	} else {
		局_创建用户名 = agent.L_agent.ID取用户名(c, 请求.Uid)
	}
	log.L_log.Log_写库存转移日志(局_新卡包.Id, 局_新卡包.NumMax, 3, 局_创建用户名, 局_角色, 局_创建用户名, 局_角色, c.ClientIP(), "自助购买")
}

func (A *AgentInventoryOld) K库存撤回(c *gin.Context) {
	var 请求 Agent库存撤回请求
	if err := c.ShouldBindJSON(&请求); err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if err := agent.L_agent.K库存撤回(c, c.GetInt("Uid"), 请求.Id, 请求.Num, 请求.Note, c.ClientIP()); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("撤回成功", c)
}

func (A *AgentInventoryOld) K库存发送(c *gin.Context) {
	var 请求 Agent库存发送请求
	if err := c.ShouldBindJSON(&请求); err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if service.NewAgentInventory(c, global.GVA_DB).Id取归属Uid(请求.SourceID) != c.GetInt("Uid") {
		response.FailWithMessage("只能将归属自己的库存,发送给别人.", c)
		return
	}
	if err := agent.L_agent.K库存发送(c, 请求.SourceID, 请求.ToUserId, 请求.Num, 请求.Note, c.ClientIP()); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("发送成功", c)
}

func (A *AgentInventoryOld) K库存延期(c *gin.Context) {
	var 请求 Agent库存撤回请求
	if err := c.ShouldBindJSON(&请求); err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if err := agent.L_agent.K库存延期(c, 请求.Id, c.GetInt("Uid"), 请求.Num); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("操作成功", c)
}

func (A *AgentInventoryOld) Get取可创建库存包列表(c *gin.Context) {
	var 请求 Agent库存单Id请求
	if err := c.ShouldBindJSON(&请求); err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}
	局_卡类树 := ka.L_ka.Q取全部可制卡类树形框列表(c, c.GetInt("Uid"))
	response.OkWithDetailed(Agent库存卡类树响应{KaClassTree: 局_卡类树}, "获取成功", c)
}

func (A *AgentInventoryOld) K库存修改备注(c *gin.Context) {
	var 请求 Agent库存撤回请求
	if err := c.ShouldBindJSON(&请求); err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if err := service.NewAgentInventory(c, global.GVA_DB).K库存修改备注(请求.Id, c.GetInt("Uid"), 请求.Note); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("操作成功", c)
}

func (A *AgentInventoryOld) Q可发送库存下级代理(c *gin.Context) {
	局_下级代理ID数组 := agent.L_agent.Q取下级代理数组(c, []int{c.GetInt("Uid")})
	if len(局_下级代理ID数组) == 0 {
		response.FailWithMessage("无直属下级代理", c)
		return
	}
	局_下级代理详情, err := service.NewUser(c, global.GVA_DB).Id取详情_数组(局_下级代理ID数组)
	if err != nil {
		response.FailWithMessage("读取失败:"+err.Error(), c)
		return
	}

	局_响应列表 := make([]Agent可发送下级, 0, len(局_下级代理详情))
	for 局_索引 := range 局_下级代理详情 {
		局_响应列表 = append(局_响应列表, Agent可发送下级{
			Id:       局_下级代理详情[局_索引].Id,
			User:     局_下级代理详情[局_索引].User,
			Disabled: 局_下级代理详情[局_索引].Status == 2,
		})
	}
	response.OkWithDetailed(局_响应列表, "成功", c)
}
