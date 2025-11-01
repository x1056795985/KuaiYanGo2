package AgentUser

import (
	"EFunc/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"server/Service/Ser_Chare"
	"server/Service/Ser_KaClass"
	"server/Service/Ser_LinkUser"
	"server/Service/Ser_Log"
	"server/Service/Ser_User"
	"server/global"
	"server/new/app/logic/common/agent"
	"server/new/app/logic/common/agentLevel"

	"server/structs/Http/response"
	DB "server/structs/db"
	. "server/utils"
	"strconv"
	"strings"
)

type Api struct{}

// GetAgentUserInfo
func (a *Api) GetAgentUserInfo(c *gin.Context) {
	var 请求 结构请求_单id
	//{"Id":2}
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	if agent.L_agent.Q取上级代理的子级代理级别(c, c.GetInt("Uid"), 请求.Id) <= 0 {
		response.FailWithMessage("只能查询自己的子级代理详细信息", c)
		return
	}

	var DB_AgentUser DB_AgentUser2

	err = global.GVA_DB.Model(DB.DB_User{}).Omit("Note", "PassWord", "SuperPassWord").Where("id = ?", 请求.Id).Find(&DB_AgentUser).Error
	// 没查到数据
	if err != nil {
		response.FailWithMessage("查询用户详细信息失败", c)
		return
	}
	//DB_AgentUser.Role = agentLevel.L_agentLevel.Q取Id代理级别(c,DB_AgentUser.Id)  //不能让下级代理知道自己是几级代理,容易三级直接想办法联系一级代理
	DB_AgentUser.UPAgentUser = agent.L_agent.ID取用户名(c, DB_AgentUser.UPAgentId)
	//DB_AgentUser.LoginAppName = Ser_AppInfo.AppId取应用名称(DB_AgentUser.LoginAppid)

	response.OkWithDetailed(DB_AgentUser, "获取成功", c)
	return
}

type DB_AgentUser2 struct {
	DB.DB_User
	LoginAppName string `json:"LoginAppName"` //登录平台App名字
	Role         int    `json:"Role"`         //
	UPAgentUser  string `json:"UPAgentUser"`  //
}

type 结构请求_单id struct {
	Id int `json:"Id"`
}

type 结构请求_GetUserList struct {
	Page     int    `json:"Page"`     // 页
	Size     int    `json:"Size"`     // 页数量
	Status   int    `json:"Status"`   // 状态id
	Type     int    `json:"Type"`     // 关键字类型  1 id 2 用户名 3绑定信息 4 动态标签
	Keywords string `json:"Keywords"` // 关键字
	Order    int    `json:"Order"`    // 0 倒序 1 正序
}

// GetUserList
// 获取用户信息列表
func (a *Api) GetAgentUserList(c *gin.Context) {
	var 请求 结构请求_GetUserList
	//{"Type":"2","Size":10,"Page":1,"Status":"1","keywords":"1"}
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	var 总数 int64
	局_所有子级代理ID := agent.L_agent.Q取下级代理数组含子级(c, []int{c.GetInt("Uid")})
	//限制只能查代理,限制只能查自己的子级代理列表
	局_DB := global.GVA_DB.Model(DB.DB_User{}).Where("UPAgentId != 0").Where("Id IN ?", 局_所有子级代理ID)

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
		case 1: //id
			局_DB.Where("Id = ?", 请求.Keywords)
		case 2: //用户名
			局_DB.Where("LOCATE(?, User)>0 ", 请求.Keywords)
		case 3: //余额大于X
			float, _ := strconv.ParseFloat(请求.Keywords, 64)
			局_DB.Where("Rmb > ?", float)
		case 4: //Email
			局_DB.Where("Email = ?", 请求.Keywords)
		case 5: //手机号
			局_DB.Where("Phone = ?", 请求.Keywords)
		case 6: //QQ
			局_DB.Where("Qq = ?", 请求.Keywords)
		}
	}

	var DB_User_简化实例 []DB_AgentUser_简化

	//Count(&总数) 必须放在where 后面 不然值会被清0
	//err = 局_DB.Count(&总数).Limit(请求.Size).Omit("login_app_name").Offset((请求.Page - 1) * 请求.Size).Find(&DB_User_简化实例).Error
	err = 局_DB.Count(&总数).Select("`db_User`.`Id`,`db_User`.`User`,`db_User`.`PassWord`,`db_User`.`Phone`,`db_User`.`Email`,`db_User`.`Qq`,`db_User`.`SuperPassWord`,`db_User`.`Status`,`db_User`.`Rmb`,`db_User`.`Note`,`db_User`.`RealNameAttestation`,`db_User`.`UPAgentId`,`db_User`.`AgentDiscount`,`db_User`.`LoginAppid`,`db_User`.`LoginIp`,`db_User`.`LoginTime`,`db_User`.`RegisterIp`,`db_User`.`RegisterTime`, (SELECT COUNT(*) FROM `db_Agent_Level` WHERE `db_Agent_Level`.`Uid` = `db_User`.`Id`) AS `Role`").
		Limit(请求.Size).
		Offset((请求.Page - 1) * 请求.Size).
		Find(&DB_User_简化实例).Error
	//	err = 局_DB.Count(&总数).Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&DB_Ka).Error
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		global.GVA_LOG.Error("GetUserList:" + err.Error())
		return
	}

	response.OkWithDetailed(结构响应_GetUserList{DB_User_简化实例, 总数}, "获取成功", c)
	return

}

type 结构响应_GetUserList struct {
	List  interface{} `json:"List"`  // 列表
	Count int64       `json:"Count"` // 总数
}

type DB_AgentUser_简化 struct {
	Id                  int     `json:"Id" gorm:"column:Id;primarykey"`                              // id
	User                string  `json:"User" gorm:"column:User;index;comment:用户登录名"`                 // 用户登录名
	Status              int     `json:"Status" gorm:"column:Status;default:1;comment:用户是状态 1正常 2冻结"` // 1正常 2冻结
	Rmb                 float64 `json:"Rmb" gorm:"column:Rmb;type:decimal(10,2);default:0;comment:余额"`
	RealNameAttestation string  `json:"RealNameAttestation" gorm:"column:RealNameAttestation;comment:实名认证,认证成功直接填写姓名未认证空"` //实名认证//认证成功直接填写姓名未认证空)
	UPAgentId           int     `json:"UPAgentId" gorm:"column:UPAgentId;comment:上级代理id"`
	AgentDiscount       float64 `json:"AgentDiscount" gorm:"column:AgentDiscount;type:decimal(10,2);default:0;comment:分成百分比"`
	LoginTime           int64   `json:"LoginTime" gorm:"column:LoginTime;comment:登录时间"`
	LoginIp             string  `json:"LoginIp" gorm:"column:LoginIp;comment:登录ip"`
	Role                int     `json:"Role" gorm:"column:Role;comment:角色"`
}

// New用户信息
func (a *Api) New代理信息(c *gin.Context) {
	var 请求 DB.DB_User
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
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
	请求.UPAgentId = c.GetInt("Uid") //下边提示不友好,直接删除
	/*	if 请求.UPAgentId != c.GetInt("Uid") {
		response.FailWithMessage("只能新增上级代理为自己的下级代理", c)
		return
	}*/

	if agentLevel.L_agentLevel.Q取Id代理级别(c, 请求.UPAgentId) >= 3 {
		response.FailWithMessage("三级代理无法发展下级代理", c)
		return
	}

	局_上级代理分成 := agent.L_agent.ID取分成百分比(c, 请求.UPAgentId)
	if 局_上级代理分成 < 请求.AgentDiscount {
		response.FailWithMessage("分成百分比最高"+strconv.Itoa(局_上级代理分成)+"%", c)
		return
	}
	局_下级代理分成 := Ser_User.Id取下级代理分成最高(请求.Id)
	if 局_下级代理分成 > 请求.AgentDiscount {
		response.FailWithMessage("该代理的下级代理已设置分成百分比为"+strconv.Itoa(局_下级代理分成)+"%,故不能设置低于该值,请联系协商", c)
		return
	}
	msg := ""
	if !utils.Z正则_校验代理用户名(请求.User, &msg) {
		response.FailWithMessage("用户名"+msg, c)
		return
	}
	_, err = Ser_User.New用户信息(请求.User, 请求.PassWord, 请求.SuperPassWord, 请求.Qq, 请求.Email, 请求.Phone, c.ClientIP(), 请求.Note, 请求.UPAgentId, 请求.AgentDiscount, 请求.Rmb, "")
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithMessage("添加成功", c)
	if 请求.Rmb != 0 {
		go Ser_Log.Log_写余额日志(请求.User, c.ClientIP(), fmt.Sprintf("管理员(%v),新增用户携带余额:%v", c.GetInt("Uid"), 请求.Rmb), 请求.Rmb)
	}
	return
}

// save 保存
func (a *Api) Save代理信息(c *gin.Context) {
	var 请求 DB.DB_User
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
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

	msg := ""
	if 请求.PassWord != "" && !utils.Z正则_校验密码(请求.PassWord, &msg) {
		response.FailWithMessage("密码"+msg, c)
		return
	}

	if 请求.Email != "" && !utils.Z正则_校验email(请求.Email, &msg) {
		response.FailWithMessage("email邮箱格式不正确", c)
		return
	}

	if 请求.SuperPassWord != "" && !utils.Z正则_校验密码(请求.SuperPassWord, &msg) {
		response.FailWithMessage("超级密码"+msg, c)
		return
	}
	// 没查到数据
	用户详情, ok := Ser_User.Id取详情(请求.Id)
	if !ok {
		response.FailWithMessage("用户不存在", c)
		return
	}
	if 用户详情.Rmb != 请求.Rmb && c.GetInt("Uid") != 1 {
		response.FailWithMessage("非系统管理员不能通过编辑改变代理余额", c)
		return
	}
	局_上级代理分成 := agent.L_agent.ID取分成百分比(c, 请求.UPAgentId)
	if 局_上级代理分成 < 请求.AgentDiscount {
		response.FailWithMessage("分成百分比最高"+strconv.Itoa(局_上级代理分成)+"%", c)
		return
	}
	局_下级代理分成 := Ser_User.Id取下级代理分成最高(请求.Id)
	if 局_下级代理分成 > 请求.AgentDiscount {
		response.FailWithMessage("该代理的下级代理已设置分成百分比为"+strconv.Itoa(局_下级代理分成)+"%,故不能设置低于该值,请联系协商", c)
		return
	}

	m := map[string]interface{}{
		"Phone":               请求.Phone,
		"Email":               请求.Email,
		"Qq":                  请求.Qq,
		"Status":              请求.Status,
		"Note":                请求.Note,
		"AgentDiscount":       请求.AgentDiscount,
		"RealNameAttestation": 请求.RealNameAttestation,
	}

	if 请求.PassWord != "" {
		m["PassWord"] = BcryptHash(请求.PassWord)
	}

	if 请求.SuperPassWord != "" {
		m["SuperPassWord"] = BcryptHash(请求.SuperPassWord)
	}

	var db = global.GVA_DB.Model(DB.DB_User{}).Where("Id= ?", 请求.Id).Updates(&m)

	if db.Error != nil {
		fmt.Printf(db.Error.Error())
		response.FailWithMessage("保存失败", c)
		return
	}
	response.OkWithMessage("保存成功"+strconv.Itoa(int(db.RowsAffected)), c)
	return
}

// Del批量修改状态
func (a *Api) Set修改状态(c *gin.Context) {
	var 请求 结构请求_批量修改状态
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
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
	err = global.GVA_DB.Model(DB.DB_User{}).Where("Id IN ? ", 请求.Id).Update("Status", 请求.Status).Error
	if err != nil {
		response.FailWithMessage("修改失败", c)
		global.GVA_LOG.Error("修改失败:" + err.Error())
		return
	}

	if 请求.Status == 2 {
		局_user数组 := make([]string, 0, len(请求.Id))
		for _, 值 := range 请求.Id {
			局_user数组 = append(局_user数组, Ser_User.Id取User(值))
		}
		_ = Ser_LinkUser.Set批量注销User数组(局_user数组, Ser_LinkUser.Z注销_管理员手动注销)

	}

	response.OkWithMessage("修改成功", c)
	return
}

type 结构请求_批量修改状态 struct {
	Id     []int `json:"Id"`     //用户id数组
	Status int   `json:"Status"` //1 解冻 2冻结
}

type 代理可制卡类授权 struct {
	KaList          []Ser_KaClass.K可制卡类授权树形框结构 `json:"KaList"`          //全部卡类列表
	IdListAuthority []int                      `json:"IdListAuthority"` //已授权卡类ID
	FunctionList    map[string]int             `json:"FunctionList"`    //可授权功能Id
	FunctionId      []int                      `json:"FunctionId"`      //已授权功能ID
}

// 获取代理可制卡类授权
func (a *Api) GetAgentKaClassAuthority(c *gin.Context) {
	var 请求 结构请求_单id
	//{"Id":2} 代理用户ID
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}
	var 局_返回 代理可制卡类授权
	局_上级代理ID := Ser_User.Id取上级代理ID(请求.Id)
	局_返回.KaList = Ser_KaClass.Q取全部可制卡类树形框列表(c, 局_上级代理ID)
	局_返回.FunctionList = agent.L_agent.Q取全部代理功能名称_MAP(c)
	var 局_可用代理功能ID数组 []int
	if 局_上级代理ID < 0 { //上级是管理员,全部功能都可以看见
		局_可用代理功能ID数组 = agent.L_agent.Q取全部代理功能ID_int数组(c)
	} else {
		_, 局_可用代理功能ID数组 = agent.L_agent.Id取代理可制卡类和可用代理功能列表(c, 局_上级代理ID)
		for key := range 局_返回.FunctionList {
			if !utils.S数组_整数是否存在(局_可用代理功能ID数组, 局_返回.FunctionList[key]) {
				delete(局_返回.FunctionList, key) //没有授权就从全部里删除
			}
		}
	}

	局_返回.IdListAuthority, 局_返回.FunctionId = agent.L_agent.Id取代理可制卡类和可用代理功能列表(c, 请求.Id)

	response.OkWithDetailed(局_返回, "获取成功", c)
	return
}

type 结构请求_设置代理可制卡类授权 struct {
	Id  int   `json:"Id"`
	KId []int `json:"KId"`
}

// 设置代理可制卡类授权
func (a *Api) SetAgentKaClassAuthority(c *gin.Context) {
	var 请求 结构请求_设置代理可制卡类授权
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	if agent.L_agent.Q取上级代理的子级代理级别(c, c.GetInt("Uid"), 请求.Id) != 1 {
		response.FailWithMessage("只能操作自己的直属下级代理", c)
		return
	}

	if utils.S数组_整数是否存在(请求.KId, DB.D代理功能_发展下级代理) && agentLevel.L_agentLevel.Q取Id代理级别(c, 请求.Id) >= 3 {
		response.FailWithMessage("该代理不可设置发展下级代理功能权限", c)
		return
	}

	var 局_本级权限 []int
	局_可制卡号, 局_功能授权 := agent.L_agent.Id取代理可制卡类和可用代理功能列表(c, c.GetInt("Uid"))

	局_本级权限 = append(局_可制卡号, 局_功能授权...) //合并功能ID和卡类ID
	局_没有卡类 := 差集(请求.KId, 局_本级权限)

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

	err = agent.L_agent.Z置Id代理可制卡类或功能授权列表(c, 请求.Id, 请求.KId)
	if err == nil {
		response.OkWithMessage("操作成功", c)
	} else {
		response.FailWithMessage("操作失败错误:"+err.Error(), c)
	}
	return
}

// 转账给下级代理
func (a *Api) SendRmbTOAgent(c *gin.Context) {
	var 请求 struct {
		Id  int     `json:"Id"`
		RMB float64 `json:"Rmb"`
	}
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	if agent.L_agent.Q取上级代理的子级代理级别(c, c.GetInt("Uid"), 请求.Id) != 1 {
		response.FailWithMessage("只能转账给自己的直属下级代理", c)
		return
	}
	var 源新余额, 目标新余额 float64
	源新余额, 目标新余额, err = Ser_User.Id余额转账(c.GetInt("Uid"), 请求.Id, 请求.RMB, c.ClientIP())
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithDetailed(map[string]float64{"sourceRmb": 源新余额, "targetRmb": 目标新余额}, "操作成功", c)
	return
}

// 差集函数，返回切片a中有但切片b中没有的元素
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

func (a *Api) Get代理组织架构图(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get代理组织架构图(c, c.GetInt("Uid")), "获取成功", c)
}
