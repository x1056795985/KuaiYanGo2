package controller

import (
	"EFunc/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_KaClass"
	"server/Service/Ser_Log"
	"server/Service/Ser_User"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/logic/common/agent"
	"server/new/app/logic/common/agentLevel"
	dbm "server/new/app/models/db"
	"server/structs/Http/response"
	DB "server/structs/db"
	utils2 "server/utils"
	"strconv"
	"strings"
)

type AgentUserFull struct {
	Common.Common
}

func NewAgentUserFullController() *AgentUserFull {
	return &AgentUserFull{}
}

type DB_AgentUser2 struct {
	DB.DB_User
	LoginAppName string `json:"loginAppName"`
	Role         int    `json:"role"`
	UPAgentUser  string `json:"uPAgentUser"`
}

type DB_AgentUser_简化 struct {
	Id                  int     `json:"id" gorm:"column:Id;primarykey"`
	User                string  `json:"user" gorm:"column:User;index;comment:用户登录名"`
	Status              int     `json:"status" gorm:"column:Status;default:1;comment:用户是状态 1正常 2冻结"`
	Rmb                 float64 `json:"rmb" gorm:"column:Rmb;type:decimal(10,2);default:0;comment:余额"`
	RealNameAttestation string  `json:"realNameAttestation" gorm:"column:RealNameAttestation;comment:实名认证"`
	UPAgentId           int     `json:"uPAgentId" gorm:"column:UPAgentId;comment:上级代理id"`
	AgentDiscount       float64 `json:"agentDiscount" gorm:"column:AgentDiscount;type:decimal(10,2);default:0;comment:分成百分比"`
	LoginTime           int64   `json:"loginTime" gorm:"column:LoginTime;comment:登录时间"`
	LoginIp             string  `json:"loginIp" gorm:"column:LoginIp;comment:登录ip"`
	Role                int     `json:"role" gorm:"column:Role;comment:角色"`
	Note                string  `json:"note" gorm:"column:Note;comment:备注"`
	Sort                int64   `json:"sort" gorm:"column:Sort;default:0;comment:排序权重"`
}

type 代理可制卡类授权 struct {
	KaList          []Ser_KaClass.K可制卡类授权树形框结构 `json:"kaList"`
	IdListAuthority []int                      `json:"idListAuthority"`
	FunctionList    map[string]int             `json:"functionList"`
	FunctionId      []int                      `json:"functionId"`
}

// Info 获取代理详情
func (C *AgentUserFull) Info(c *gin.Context) {
	var 请求 struct {
		Id int `json:"id"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	var DB_AgentUser DB_AgentUser2
	err := global.GVA_DB.Model(DB.DB_User{}).Omit("PassWord", "SuperPassWord").Where("id = ?", 请求.Id).Find(&DB_AgentUser).Error
	if err != nil {
		response.FailWithMessage("查询用户详细信息失败", c)
		return
	}
	DB_AgentUser.Role = agentLevel.L_agentLevel.Q取Id代理级别(c, DB_AgentUser.Id)
	DB_AgentUser.UPAgentUser = agent.L_agent.ID取用户名(c, DB_AgentUser.UPAgentId)
	DB_AgentUser.LoginAppName = Ser_AppInfo.AppId取应用名称(DB_AgentUser.LoginAppid)
	response.OkWithDetailed(DB_AgentUser, "获取成功", c)
}

// GetList 获取代理列表
func (C *AgentUserFull) GetList(c *gin.Context) {
	var 请求 struct {
		Page     int    `json:"page"`
		Size     int    `json:"size"`
		Status   int    `json:"status"`
		Type     int    `json:"type"`
		Keywords string `json:"keywords"`
		Order    int    `json:"order"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	var 总数 int64
	局_DB := global.GVA_DB.Model(DB.DB_User{}).Where("UPAgentId != 0")
	if 请求.Order == 1 {
		局_DB.Order("Sort DESC,Id ASC")
	} else {
		局_DB.Order("Sort ASC,Id DESC")
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
			float, _ := strconv.ParseFloat(请求.Keywords, 64)
			局_DB.Where("Rmb > ?", float)
		case 4:
			局_DB.Where("Email = ?", 请求.Keywords)
		case 5:
			局_DB.Where("Phone = ?", 请求.Keywords)
		case 6:
			局_DB.Where("Qq = ?", 请求.Keywords)
		}
	}

	var DB_User_简化实例 []DB_AgentUser_简化
	err := 局_DB.Count(&总数).Select("`db_User`.`Id`,`db_User`.`User`,`db_User`.`PassWord`,`db_User`.`Phone`,`db_User`.`Email`,`db_User`.`Qq`,`db_User`.`SuperPassWord`,`db_User`.`Status`,`db_User`.`Rmb`,`db_User`.`Note`,`db_User`.`RealNameAttestation`,`db_User`.`UPAgentId`,`db_User`.`AgentDiscount`,`db_User`.`LoginAppid`,`db_User`.`LoginIp`,`db_User`.`LoginTime`,`db_User`.`RegisterIp`,`db_User`.`Sort`,`db_User`.`RegisterTime`, (SELECT COUNT(*) FROM `db_Agent_Level` WHERE `db_Agent_Level`.`Uid` = `db_User`.`Id`) AS `Role`").
		Where("UPAgentId != 0").
		Order("Id DESC").
		Limit(请求.Size).
		Offset((请求.Page - 1) * 请求.Size).
		Find(&DB_User_简化实例).Error
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		return
	}
	response.OkWithDetailed(struct {
		List  interface{} `json:"list"`
		Count int64       `json:"count"`
	}{DB_User_简化实例, 总数}, "获取成功", c)
}

// New 新建代理
func (C *AgentUserFull) New(c *gin.Context) {
	var 请求 DB.DB_User
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
	if agent.L_agent.ID取用户名(c, 请求.UPAgentId) == "" {
		response.FailWithMessage("上级代理不存在", c)
		return
	}
	局_上级代理分成 := agent.L_agent.ID取分成百分比(c, 请求.UPAgentId)
	if 局_上级代理分成 < int(请求.AgentDiscount) {
		response.FailWithMessage("分成百分比最高"+strconv.Itoa(局_上级代理分成)+"%", c)
		return
	}
	局_下级代理分成 := Ser_User.Id取下级代理分成最高(请求.Id)
	if 局_下级代理分成 > int(请求.AgentDiscount) {
		response.FailWithMessage("该代理的下级代理已设置分成百分比为"+strconv.Itoa(局_下级代理分成)+"%,故不能设置低于该值,请联系协商", c)
		return
	}
	msg := ""
	if !utils.Z正则_校验代理用户名(请求.User, &msg) {
		response.FailWithMessage("用户名"+msg, c)
		return
	}
	_, err := Ser_User.New用户信息(请求.User, 请求.PassWord, 请求.SuperPassWord, 请求.Qq, 请求.Email, 请求.Phone, c.ClientIP(), 请求.Note, 请求.UPAgentId, 请求.AgentDiscount, 请求.Rmb, "")
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("添加成功", c)
	if 请求.Rmb != 0 {
		go Ser_Log.Log_写余额日志(请求.User, c.ClientIP(), fmt.Sprintf("管理员(%v),新增用户携带余额:%v", c.GetInt("Uid"), 请求.Rmb), 请求.Rmb)
	}
}

// Save 保存代理信息
func (C *AgentUserFull) Save(c *gin.Context) {
	var 请求 DB.DB_User
	if !C.ToJSON(c, &请求) {
		return
	}
	if 请求.Id <= 0 {
		response.FailWithMessage("Id错误", c)
		return
	}
	msg := ""
	if 请求.PassWord != "" && !utils.Z正则_校验密码(请求.PassWord, &msg) {
		response.FailWithMessage("密码"+msg, c)
		return
	}
	if 请求.Email != "" && !utils.Z正则_校验email(请求.Email, &msg) {
		response.FailWithMessage("email邮箱格式不正确", c)
		return
	}
	if 请求.SuperPassWord != "" {
		response.FailWithMessage("超级密码"+msg, c)
		return
	}
	用户详情, ok := Ser_User.Id取详情(请求.Id)
	if !ok {
		response.FailWithMessage("用户不存在", c)
		return
	}
	if 用户详情.Rmb != 请求.Rmb && c.GetInt("Uid") != 1 {
		response.FailWithMessage("非系统管理员不能通过编辑改变代理余额", c)
		return
	}
	if agent.L_agent.ID取用户名(c, 请求.UPAgentId) == "" {
		response.FailWithMessage("上级代理不存在", c)
		return
	}
	if agentLevel.L_agentLevel.Q取Id代理级别(c, 请求.UPAgentId) >= 3 {
		response.FailWithMessage("上级代理为三级代理无法发展下级代理", c)
		return
	}
	局_上级代理分成 := agent.L_agent.ID取分成百分比(c, 请求.UPAgentId)
	if 局_上级代理分成 < int(请求.AgentDiscount) {
		response.FailWithMessage("分成百分比最高"+strconv.Itoa(局_上级代理分成)+"%", c)
		return
	}
	局_下级代理分成 := Ser_User.Id取下级代理分成最高(请求.Id)
	if 局_下级代理分成 > int(请求.AgentDiscount) {
		response.FailWithMessage("该代理的下级代理已设置分成百分比为"+strconv.Itoa(局_下级代理分成)+"%,故不能设置低于该值,请联系协商", c)
		return
	}

	m := map[string]interface{}{
		"Phone":               请求.Phone,
		"Email":               请求.Email,
		"Qq":                  请求.Qq,
		"Status":              请求.Status,
		"Rmb":                 请求.Rmb,
		"Note":                请求.Note,
		"AgentDiscount":       请求.AgentDiscount,
		"RealNameAttestation": 请求.RealNameAttestation,
	}
	if 请求.PassWord != "" {
		m["PassWord"] = utils2.BcryptHash(请求.PassWord)
	}
	if 请求.SuperPassWord != "" {
		m["SuperPassWord"] = utils2.BcryptHash(请求.SuperPassWord)
	}

	var db = global.GVA_DB.Model(DB.DB_User{}).Where("Id= ?", 请求.Id).Updates(&m)
	if db.Error != nil {
		fmt.Printf(db.Error.Error())
		response.FailWithMessage("保存失败", c)
		return
	}
	if 用户详情.Rmb != 请求.Rmb {
		go Ser_Log.Log_写余额日志(用户详情.User, c.ClientIP(), "管理员ID:"+strconv.Itoa(c.GetInt("Uid"))+"编辑用户信息余额变化:"+utils.Float64到文本(用户详情.Rmb, 2)+"=>"+utils.Float64到文本(请求.Rmb, 2), 请求.Rmb-用户详情.Rmb)
	}
	response.OkWithMessage("保存成功"+strconv.Itoa(int(db.RowsAffected)), c)
}

// SetStatus 批量修改代理状态
func (C *AgentUserFull) SetStatus(c *gin.Context) {
	var 请求 struct {
		Id     []int `json:"id"`
		Status int   `json:"status"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	if len(请求.Id) == 0 {
		response.FailWithMessage("Id数组为空", c)
		return
	}
	if 请求.Status != 1 && 请求.Status != 2 {
		response.FailWithMessage("修改失败:Status状态代码错误", c)
		return
	}

	var err error
	if 请求.Status == 2 {
		err = global.GVA_DB.Model(DB.DB_User{}).Where("Id IN ? ", 请求.Id).Update("Status", 2).Error
	} else {
		err = global.GVA_DB.Model(DB.DB_User{}).Where("Id IN ? ", 请求.Id).Update("Status", 1).Error
	}
	if err != nil {
		response.FailWithMessage("修改失败", c)
		return
	}
	response.OkWithMessage("修改成功", c)
}

// Delete 批量删除代理
func (C *AgentUserFull) Delete(c *gin.Context) {
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
	局_子级代理ID数组 := agent.L_agent.Q取下级代理数组含子级(c, 请求.Id)
	if len(局_子级代理ID数组) > 0 {
		response.FailWithMessage("用户有子级代理,暂不可删除,请先根据代理组织结构图,删除所有子级代理后,再删除该用户", c)
		return
	}

	err := agent.L_agent.S删除代理(c, 请求.Id)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetAgentKaClassAuthority 获取代理可制卡类授权
func (C *AgentUserFull) GetAgentKaClassAuthority(c *gin.Context) {
	var 请求 struct {
		Id int `json:"id"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	var 局_返回 代理可制卡类授权
	局_上级代理ID := Ser_User.Id取上级代理ID(请求.Id)
	局_返回.KaList = Ser_KaClass.Q取全部可制卡类树形框列表(c, 局_上级代理ID)
	局_返回.FunctionList = agent.L_agent.Q取全部代理功能名称_MAP(c)
	var 局_可用代理功能ID数组 []int
	if 局_上级代理ID < 0 {
		局_可用代理功能ID数组 = agent.L_agent.Q取全部代理功能ID_int数组(c)
	} else {
		_, 局_可用代理功能ID数组 = agent.L_agent.Id取代理可制卡类和可用代理功能列表(c, 局_上级代理ID)
		for key := range 局_返回.FunctionList {
			if !utils.S数组_整数是否存在(局_可用代理功能ID数组, 局_返回.FunctionList[key]) {
				delete(局_返回.FunctionList, key)
			}
		}
	}
	局_返回.IdListAuthority, 局_返回.FunctionId = agent.L_agent.Id取代理可制卡类和可用代理功能列表(c, 请求.Id)
	response.OkWithDetailed(局_返回, "获取成功", c)
}

// SetAgentKaClassAuthority 设置代理可制卡类授权
func (C *AgentUserFull) SetAgentKaClassAuthority(c *gin.Context) {
	var 请求 struct {
		Id  int   `json:"id"`
		KId []int `json:"kId"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	if utils.S数组_整数是否存在(请求.KId, DB.D代理功能_发展下级代理) && agentLevel.L_agentLevel.Q取Id代理级别(c, 请求.Id) >= 3 {
		response.FailWithMessage("三级代理不可设置发展下级代理功能", c)
		return
	}
	var 局_已有卡类 []int
	global.GVA_DB.Model(dbm.DB_KaClass{}).Select("Id").Where("Id IN ?", 请求.KId).Find(&局_已有卡类)
	局_上级代理ID := Ser_User.Id取上级代理ID(请求.Id)
	var 局_可用功能列表 []int
	if 局_上级代理ID < 0 {
		局_可用功能列表 = agent.L_agent.Q取全部代理功能ID_int数组(c)
	} else {
		_, 局_可用功能列表 = agent.L_agent.Id取代理可制卡类和可用代理功能列表(c, 局_上级代理ID)
	}
	局_已有卡类 = append(局_已有卡类, 局_可用功能列表...)
	局_没有卡类 := 差集(请求.KId, 局_已有卡类)
	if len(局_没有卡类) > 0 {
		strSlice := make([]string, len(局_没有卡类))
		var 局_可授权功能Map = agent.L_agent.Q取全部代理功能ID_MAP(c)
		for i, num := range 局_没有卡类 {
			if num < 0 {
				strSlice[i] = 局_可授权功能Map[num]
			} else {
				strSlice[i] = strconv.Itoa(num)
			}
		}
		response.FailWithMessage("有不可选中卡类ID或功能:"+strings.Join(strSlice, ","), c)
		return
	}

	err := agent.L_agent.Z置Id代理可制卡类或功能授权列表(c, 请求.Id, 请求.KId)
	if err == nil {
		response.OkWithMessage("操作成功", c)
	} else {
		response.FailWithMessage("操作失败错误:"+err.Error(), c)
	}
}

func 差集(a, b []int) []int {
	m := make(map[int]bool)
	for _, v := range b {
		m[v] = true
	}
	var 结果 []int
	for _, v := range a {
		if !m[v] {
			结果 = append(结果, v)
		}
	}
	return 结果
}
