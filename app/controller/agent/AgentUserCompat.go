package controller

import (
	"EFunc/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"server/app/global"
	"server/app/logic/admin/L_chart"
	"server/app/logic/common/agent"
	"server/app/logic/common/agentLevel"
	"server/app/logic/common/ka"
	"server/app/logic/common/log"
	"server/app/logic/common/user"
	"server/app/models/constant"
	"server/app/models/db"
	"server/app/models/old/response"
	"server/app/service"
	utils2 "server/app/utils"
	"strconv"
	"strings"
)

type Agent用户详情 struct {
	db.DB_User
	LoginAppName string `json:"loginAppName"`
	Role         int    `json:"role"`
	UPAgentUser  string `json:"upAgentUser"`
}

type Agent用户列表请求 struct {
	Page     int    `json:"page"`
	Size     int    `json:"size"`
	Status   int    `json:"status"`
	Type     int    `json:"type"`
	Keywords string `json:"keywords"`
	Order    int    `json:"order"`
}

type Agent用户列表项 struct {
	Id                  int     `json:"id" gorm:"column:Id;primarykey"`
	User                string  `json:"user" gorm:"column:User;index;comment:用户登录名"`
	Status              int     `json:"status" gorm:"column:Status;default:1;comment:用户是状态 1正常 2冻结"`
	Rmb                 float64 `json:"rmb" gorm:"column:Rmb;type:decimal(10,2);default:0;comment:余额"`
	RealNameAttestation string  `json:"realNameAttestation" gorm:"column:RealNameAttestation;comment:实名认证,认证成功直接填写姓名未认证空"`
	UPAgentId           int     `json:"upAgentId" gorm:"column:UPAgentId;comment:上级代理id"`
	AgentDiscount       float64 `json:"agentDiscount" gorm:"column:AgentDiscount;type:decimal(10,2);default:0;comment:分成百分比"`
	LoginTime           int64   `json:"loginTime" gorm:"column:LoginTime;comment:登录时间"`
	LoginIp             string  `json:"loginIp" gorm:"column:LoginIp;comment:登录ip"`
	Role                int     `json:"role" gorm:"column:Role;comment:角色"`
}

type Agent列表响应 struct {
	List  interface{} `json:"list"`
	Count int64       `json:"count"`
}

type Agent批量状态请求 struct {
	Id     []int `json:"id"`
	Status int   `json:"status"`
}

type Agent权限响应 struct {
	KaList          []ka.K可制卡类授权树形框结构 `json:"kaList"`
	IdListAuthority []int             `json:"idListAuthority"`
	FunctionList    map[string]int    `json:"functionList"`
	FunctionId      []int             `json:"functionId"`
}

type Agent授权请求 struct {
	Id  int   `json:"id"`
	KId []int `json:"kId"`
}

func (C *AgentUser) GetAgentUserInfo(c *gin.Context) {
	var 请求 struct {
		Id int `json:"id"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	if agent.L_agent.Q取上级代理的子级代理级别(c, c.GetInt("Uid"), 请求.Id) <= 0 {
		response.FailWithMessage("只能查询自己的子级代理详细信息", c)
		return
	}

	var 局_用户详情 Agent用户详情
	if err := global.GVA_DB.Model(db.DB_User{}).
		Omit("Note", "PassWord", "SuperPassWord").
		Where("id = ?", 请求.Id).
		Find(&局_用户详情).Error; err != nil {
		response.FailWithMessage("查询用户详细信息失败", c)
		return
	}

	局_用户详情.UPAgentUser = agent.L_agent.ID取用户名(c, 局_用户详情.UPAgentId)
	response.OkWithDetailed(局_用户详情, "获取成功", c)
}

func (C *AgentUser) GetAgentUserList(c *gin.Context) {
	var 请求 Agent用户列表请求
	if !C.ToJSON(c, &请求) {
		return
	}

	局_所有子级代理ID := agent.L_agent.Q取下级代理数组含子级(c, []int{c.GetInt("Uid")})
	局_DB := global.GVA_DB.Model(db.DB_User{}).Where("UPAgentId != 0").Where("Id IN ?", 局_所有子级代理ID)

	if 请求.Order == 1 {
		局_DB.Order("Id ASC")
	} else {
		局_DB.Order("Id DESC")
	}
	if 请求.Status == 1 || 请求.Status == 2 {
		局_DB.Where("Status = ?", 请求.Status)
	}
	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1:
			局_DB.Where("Id = ?", 请求.Keywords)
		case 2:
			局_DB.Where("LOCATE(?, User)>0 ", 请求.Keywords)
		case 3:
			局_余额, _ := strconv.ParseFloat(请求.Keywords, 64)
			局_DB.Where("Rmb > ?", 局_余额)
		case 4:
			局_DB.Where("Email = ?", 请求.Keywords)
		case 5:
			局_DB.Where("Phone = ?", 请求.Keywords)
		case 6:
			局_DB.Where("Qq = ?", 请求.Keywords)
		}
	}

	var 局_总数 int64
	var 局_列表 []Agent用户列表项
	err := 局_DB.Count(&局_总数).
		Select("`db_User`.`Id`,`db_User`.`User`,`db_User`.`PassWord`,`db_User`.`Phone`,`db_User`.`Email`,`db_User`.`Qq`,`db_User`.`SuperPassWord`,`db_User`.`Status`,`db_User`.`Rmb`,`db_User`.`Note`,`db_User`.`RealNameAttestation`,`db_User`.`UPAgentId`,`db_User`.`AgentDiscount`,`db_User`.`LoginAppid`,`db_User`.`LoginIp`,`db_User`.`LoginTime`,`db_User`.`RegisterIp`,`db_User`.`RegisterTime`, (SELECT COUNT(*) FROM `db_Agent_Level` WHERE `db_Agent_Level`.`Uid` = `db_User`.`Id`) AS `Role`").
		Limit(请求.Size).
		Offset((请求.Page - 1) * 请求.Size).
		Find(&局_列表).Error
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		global.GVA_LOG.Println("GetUserList:" + err.Error())
		return
	}

	response.OkWithDetailed(Agent列表响应{List: 局_列表, Count: 局_总数}, "获取成功", c)
}

func (C *AgentUser) New代理信息(c *gin.Context) {
	var 请求 db.DB_User
	if !C.ToJSON(c, &请求) {
		return
	}
	if 请求.Id != 0 {
		response.FailWithMessage("添加代理不能有id值", c)
		return
	}
	if 请求.Rmb != 0 && c.GetInt("Uid") != 1 {
		response.FailWithMessage("非系统管理员只能创建余额=0的代理用户", c)
		return
	}

	请求.UPAgentId = c.GetInt("Uid")
	if agentLevel.L_agentLevel.Q取Id代理级别(c, 请求.UPAgentId) >= 3 {
		response.FailWithMessage("三级代理无法发展下级代理", c)
		return
	}

	局_上级代理分成 := agent.L_agent.ID取分成百分比(c, 请求.UPAgentId)
	if 局_上级代理分成 < 请求.AgentDiscount {
		response.FailWithMessage("分成百分比最高"+strconv.Itoa(局_上级代理分成)+"%", c)
		return
	}
	局_下级代理分成 := service.NewUser(c, global.GVA_DB).Id取下级代理分成最高(请求.Id)
	if 局_下级代理分成 > 请求.AgentDiscount {
		response.FailWithMessage("该代理的下级代理已设置分成百分比为"+strconv.Itoa(局_下级代理分成)+"%,故不能设置低于该值,请联系协商", c)
		return
	}

	var 局_错误文本 string
	if !utils.Z正则_校验代理用户名(请求.User, &局_错误文本) {
		response.FailWithMessage("用户名"+局_错误文本, c)
		return
	}

	if _, err := user.L_user.New用户信息(c, 请求.User, 请求.PassWord, 请求.SuperPassWord, 请求.Qq, 请求.Email, 请求.Phone, c.ClientIP(), 请求.Note, 请求.UPAgentId, 请求.AgentDiscount, 请求.Rmb, ""); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithMessage("添加成功", c)
	if 请求.Rmb != 0 {
		go log.L_log.Log_写余额日志(请求.User, c.ClientIP(), fmt.Sprintf("管理员(%v),新增用户携带余额:%v", c.GetInt("Uid"), 请求.Rmb), 请求.Rmb)
	}
}

func (C *AgentUser) Save代理信息(c *gin.Context) {
	var 请求 db.DB_User
	if !C.ToJSON(c, &请求) {
		return
	}
	if 请求.Id <= 0 {
		response.FailWithMessage("Id错误", c)
		return
	}
	if agent.L_agent.Q取上级代理的子级代理级别(c, c.GetInt("Uid"), 请求.Id) != 1 {
		response.FailWithMessage("权限不足,只能操作自己的直属子级代理", c)
		return
	}

	var 局_错误文本 string
	if 请求.PassWord != "" && !utils.Z正则_校验密码(请求.PassWord, &局_错误文本) {
		response.FailWithMessage("密码"+局_错误文本, c)
		return
	}
	if 请求.Email != "" && !utils.Z正则_校验email(请求.Email, &局_错误文本) {
		response.FailWithMessage("email邮箱格式不正确", c)
		return
	}
	if 请求.SuperPassWord != "" && !utils.Z正则_校验密码(请求.SuperPassWord, &局_错误文本) {
		response.FailWithMessage("超级密码"+局_错误文本, c)
		return
	}

	局_用户详情, ok := service.NewUser(c, global.GVA_DB).Id取详情(请求.Id)
	if !ok {
		response.FailWithMessage("用户不存在", c)
		return
	}
	if 局_用户详情.Rmb != 请求.Rmb && c.GetInt("Uid") != 1 {
		response.FailWithMessage("非系统管理员不能通过编辑改变代理余额", c)
		return
	}

	局_上级代理分成 := agent.L_agent.ID取分成百分比(c, 请求.UPAgentId)
	if 局_上级代理分成 < 请求.AgentDiscount {
		response.FailWithMessage("分成百分比最高"+strconv.Itoa(局_上级代理分成)+"%", c)
		return
	}
	局_下级代理分成 := service.NewUser(c, global.GVA_DB).Id取下级代理分成最高(请求.Id)
	if 局_下级代理分成 > 请求.AgentDiscount {
		response.FailWithMessage("该代理的下级代理已设置分成百分比为"+strconv.Itoa(局_下级代理分成)+"%,故不能设置低于该值,请联系协商", c)
		return
	}

	局_更新字段 := map[string]interface{}{
		"Phone":               请求.Phone,
		"Email":               请求.Email,
		"Qq":                  请求.Qq,
		"Status":              请求.Status,
		"Note":                请求.Note,
		"AgentDiscount":       请求.AgentDiscount,
		"RealNameAttestation": 请求.RealNameAttestation,
	}
	if 请求.PassWord != "" {
		局_更新字段["PassWord"] = utils2.BcryptHash(请求.PassWord)
	}
	if 请求.SuperPassWord != "" {
		局_更新字段["SuperPassWord"] = utils2.BcryptHash(请求.SuperPassWord)
	}

	局_DB := global.GVA_DB.Model(db.DB_User{}).Where("Id= ?", 请求.Id).Updates(&局_更新字段)
	if 局_DB.Error != nil {
		response.FailWithMessage("保存失败", c)
		return
	}
	response.OkWithMessage("保存成功"+strconv.Itoa(int(局_DB.RowsAffected)), c)
}

func (C *AgentUser) Set修改状态(c *gin.Context) {
	var 请求 Agent批量状态请求
	if !C.ToJSON(c, &请求) {
		return
	}
	if len(请求.Id) == 0 {
		response.FailWithMessage("Id数组为空", c)
		return
	}
	if !agent.L_agent.S是否都为子级代理(c, c.GetInt("Uid"), 请求.Id) {
		response.FailWithMessage("权限不足,只能操作自己的子级代理", c)
		return
	}
	if 请求.Status != 1 && 请求.Status != 2 {
		response.FailWithMessage("修改失败:Status状态代码错误", c)
		return
	}

	if err := global.GVA_DB.Model(db.DB_User{}).Where("Id IN ? ", 请求.Id).Update("Status", 请求.Status).Error; err != nil {
		response.FailWithMessage("修改失败", c)
		global.GVA_LOG.Println("修改失败:" + err.Error())
		return
	}

	if 请求.Status == 2 {
		局_User数组 := make([]string, 0, len(请求.Id))
		for _, 局_Id := range 请求.Id {
			局_User数组 = append(局_User数组, service.NewUser(c, global.GVA_DB).Id取User(局_Id))
		}
		_ = service.NewLinksToken(c, global.GVA_DB).Set批量注销User数组(局_User数组, constant.Z注销_管理员手动注销)
	}

	response.OkWithMessage("修改成功", c)
}

func (C *AgentUser) GetAgentKaClassAuthority(c *gin.Context) {
	var 请求 struct {
		Id int `json:"id"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	var 局_返回 Agent权限响应
	局_上级代理ID := service.NewUser(c, global.GVA_DB).Id取上级代理ID(请求.Id)
	局_返回.KaList = ka.L_ka.Q取全部可制卡类树形框列表(c, 局_上级代理ID)
	局_返回.FunctionList = agent.L_agent.Q取全部代理功能名称_MAP(c)

	var 局_可用代理功能ID数组 []int
	if 局_上级代理ID < 0 {
		局_可用代理功能ID数组 = agent.L_agent.Q取全部代理功能ID_int数组(c)
	} else {
		_, 局_可用代理功能ID数组 = agent.L_agent.Id取代理可制卡类和可用代理功能列表(c, 局_上级代理ID)
		for 局_名称 := range 局_返回.FunctionList {
			if !utils.S数组_整数是否存在(局_可用代理功能ID数组, 局_返回.FunctionList[局_名称]) {
				delete(局_返回.FunctionList, 局_名称)
			}
		}
	}

	局_返回.IdListAuthority, 局_返回.FunctionId = agent.L_agent.Id取代理可制卡类和可用代理功能列表(c, 请求.Id)
	response.OkWithDetailed(局_返回, "获取成功", c)
}

func (C *AgentUser) SetAgentKaClassAuthority(c *gin.Context) {
	var 请求 Agent授权请求
	if !C.ToJSON(c, &请求) {
		return
	}
	if agent.L_agent.Q取上级代理的子级代理级别(c, c.GetInt("Uid"), 请求.Id) != 1 {
		response.FailWithMessage("只能操作自己的直属下级代理", c)
		return
	}
	if utils.S数组_整数是否存在(请求.KId, db.D代理功能_发展下级代理) && agentLevel.L_agentLevel.Q取Id代理级别(c, 请求.Id) >= 3 {
		response.FailWithMessage("该代理不可设置发展下级代理功能权限", c)
		return
	}

	局_可制卡号, 局_功能授权 := agent.L_agent.Id取代理可制卡类和可用代理功能列表(c, c.GetInt("Uid"))
	局_本级权限 := append(局_可制卡号, 局_功能授权...)
	局_缺失权限 := C.取权限差集(请求.KId, 局_本级权限)
	if len(局_缺失权限) > 0 {
		局_名称切片 := make([]string, len(局_缺失权限))
		局_功能Map := agent.L_agent.Q取全部代理功能ID_MAP(c)
		for 局_索引, 局_值 := range 局_缺失权限 {
			if 局_值 < 0 {
				局_名称切片[局_索引] = 局_功能Map[局_值]
			} else {
				局_名称切片[局_索引] = strconv.Itoa(局_值)
			}
		}
		response.FailWithMessage("有不可选中卡类ID或功能:"+strings.Join(局_名称切片, ","), c)
		return
	}

	if err := agent.L_agent.Z置Id代理可制卡类或功能授权列表(c, 请求.Id, 请求.KId); err != nil {
		response.FailWithMessage("操作失败错误:"+err.Error(), c)
		return
	}
	response.OkWithMessage("操作成功", c)
}

func (C *AgentUser) SendRmbTOAgent(c *gin.Context) {
	var 请求 struct {
		Id  int     `json:"id"`
		Rmb float64 `json:"rmb"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	if agent.L_agent.Q取上级代理的子级代理级别(c, c.GetInt("Uid"), 请求.Id) != 1 {
		response.FailWithMessage("只能转账给自己的直属下级代理", c)
		return
	}

	err := user.L_user.Id余额转账(c, c.GetInt("Uid"), 请求.Id, 请求.Rmb)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("操作成功", c)
}

func (C *AgentUser) Get代理组织架构图(c *gin.Context) {
	response.OkWithDetailed(L_chart.Get代理组织架构图(c, c.GetInt("Uid")), "获取成功", c)
}

func (C *AgentUser) Delete代理用户(c *gin.Context) {
	var 请求 struct {
		Id []int `json:"id"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	if len(请求.Id) == 0 {
		response.FailWithMessage("Id数组为空", c)
		return
	}
	if !agent.L_agent.S是否都为子级代理(c, c.GetInt("Uid"), 请求.Id) {
		response.FailWithMessage("权限不足,只能删除自己的子级代理", c)
		return
	}

	局_子级代理ID数组 := agent.L_agent.Q取下级代理数组含子级(c, 请求.Id)
	if len(局_子级代理ID数组) > 0 {
		response.FailWithMessage("用户有子级代理,暂不可删除,请先删除所有子级代理后再试", c)
		return
	}
	if err := agent.L_agent.S删除代理(c, 请求.Id); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (C *AgentUser) 取权限差集(a []int, b []int) []int {
	局_Map := make(map[int]bool)
	for _, 局_值 := range b {
		局_Map[局_值] = true
	}
	局_结果 := make([]int, 0, len(a))
	for _, 局_值 := range a {
		if !局_Map[局_值] {
			局_结果 = append(局_结果, 局_值)
		}
	}
	return 局_结果
}
