package controller

import (
	"EFunc/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/Service/Ser_Admin"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_LinkUser"
	"server/Service/Ser_Log"
	"server/Service/Ser_User"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/logic/common/agent"
	"server/new/app/logic/common/agentLevel"
	"server/new/app/logic/common/setting"
	"server/structs/Http/response"
	DB "server/structs/db"
	utils2 "server/utils"
	"strconv"
)

type UserCtrl struct {
	Common.Common
}

func NewUserController() *UserCtrl {
	return &UserCtrl{}
}

type 响应_GetAdminInfo struct {
	AdminInfo     DB.DB_Admin `json:"adminInfo"`
	UserMsgNoRead int64       `json:"userMsgNoRead"`
	ServerName    string      `json:"serverName"`
}

type DB_User2 struct {
	DB.DB_User
	LoginAppName string `json:"loginAppName"`
	Role         int    `json:"role"`
}

type DB_User_简化 struct {
	Id                  int     `json:"id" gorm:"column:Id;primarykey"`
	User                string  `json:"user" gorm:"column:User;index;comment:用户登录名"`
	Status              int     `json:"status" gorm:"column:Status;default:1;comment:用户是状态 1正常 2冻结"`
	Rmb                 float64 `json:"rmb" gorm:"column:Rmb;type:decimal(10,2);default:0;comment:余额"`
	RealNameAttestation string  `json:"realNameAttestation" gorm:"column:RealNameAttestation;comment:实名认证"`
	LoginAppid          string  `json:"loginAppid" gorm:"column:LoginAppid;comment:最后登录appid"`
	LoginAppName        string  `json:"loginAppName"`
	LoginIp             string  `json:"loginIp" gorm:"column:LoginIp;comment:登录ip"`
	LoginTime           int64   `json:"loginTime" gorm:"column:LoginTime;comment:登录时间"`
	RegisterIp          string  `json:"registerIp" gorm:"column:RegisterIp;comment:注册ip"`
	RegisterTime        int64   `json:"registerTime" gorm:"column:RegisterTime;comment:注册时间"`
	UPAgentId           int     `json:"uPAgentId" gorm:"column:UPAgentId;comment:上级代理id"`
	AgentDiscount       int     `json:"agentDiscount" gorm:"column:AgentDiscount;default:0;comment:分成百分比"`
	Note                string  `json:"note" gorm:"column:Note;comment:备注"`
}

type 请求_UserSetStatus struct {
	Id     []int `json:"id"`
	Status int   `json:"status"`
}

type 请求_UserBatchRMB struct {
	Id   []int   `json:"id"`
	RMB  float64 `json:"rMB"`
	Note string  `json:"note"`
}

// GetAdminInfo 获取当前管理员信息
func (C *UserCtrl) GetAdminInfo(c *gin.Context) {
	Uid := c.GetInt("Uid")
	var DB_user DB.DB_Admin
	err := global.GVA_DB.Model(DB.DB_Admin{}).Where("id = ?", Uid).First(&DB_user).Error
	if err != nil {
		response.FailWithMessage("查询失败", c)
		global.GVA_LOG.Error("Uid:" + strconv.Itoa(Uid) + "GetUserInfo错误:" + err.Error())
		return
	}

	局_未读用户消息数量 := Ser_Log.Y用户消息_取未读数量()
	response.OkWithDetailed(响应_GetAdminInfo{
		AdminInfo:     DB_user,
		UserMsgNoRead: 局_未读用户消息数量,
		ServerName:    setting.Q系统设置().X系统名称,
	}, "获取成功", c)
}

// OutLogin 管理员退出登录
func (C *UserCtrl) OutLogin(c *gin.Context) {
	err := Ser_LinkUser.Set批量注销Uid(c.GetInt("Uid"), Ser_LinkUser.Z注销_用户操作注销)
	if err != nil {
		response.FailWithMessage("注销失败", c)
		return
	}
	response.OkWithMessage("注销成功", c)
}

// AdminNewPassword 修改管理员密码
func (C *UserCtrl) AdminNewPassword(c *gin.Context) {
	var 请求 struct {
		NewPassword string `json:"newPassword"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	var msg = ""
	if !utils.Z正则_校验密码(请求.NewPassword, &msg) {
		response.FailWithMessage("密码"+msg, c)
		return
	}
	Uid := c.GetInt("Uid")
	err := Ser_Admin.Id置新密码(Uid, 请求.NewPassword)
	if err != nil {
		response.FailWithMessage("修改失败", c)
		return
	}
	response.OkWithMessage("修改成功", c)
}

// GetUserInfo 获取用户详情
func (C *UserCtrl) GetUserInfo(c *gin.Context) {
	var 请求 struct {
		Id int `json:"id"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	var DB_user DB_User2
	err := global.GVA_DB.Model(DB.DB_User{}).Omit("PassWord", "SuperPassWord").Where("id = ?", 请求.Id).Find(&DB_user).Error
	if err != nil {
		response.FailWithMessage("查询用户详细信息失败", c)
		return
	}
	DB_user.Role = agentLevel.L_agentLevel.Q取Id代理级别(c, DB_user.Id)
	if DB_user.LoginAppid != 0 {
		AppName := ""
		_ = global.GVA_DB.Model(DB.DB_AppInfo{}).Select("AppName").Where("AppId = ?", DB_user.LoginAppid).First(&AppName).Error
		DB_user.LoginAppName = AppName
	}
	response.OkWithDetailed(DB_user, "获取成功", c)
}

// GetUserList 获取用户列表
func (C *UserCtrl) GetUserList(c *gin.Context) {
	var 请求 struct {
		Page     int    `json:"page"`
		Size     int    `json:"size"`
		Status   int    `json:"status"`
		Type     int    `json:"type"`
		Keywords string `json:"keywords"`
		Order    int    `json:"order"`
		Role     int    `json:"role"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	var 总数 int64
	局_DB := global.GVA_DB.Model(DB.DB_User{})

	局_排序 := map[int]string{0: "Id ASC", 1: "Id DESC", 2: "Id ASC", 3: "LoginTime DESC", 4: "LoginTime ASC"}
	if utils.Map_键名是否存在(局_排序, 请求.Order) {
		局_DB.Order(局_排序[请求.Order])
	}
	if 请求.Status == 1 || 请求.Status == 2 {
		局_DB.Where("Status = ?", 请求.Status)
	}
	switch 请求.Role {
	case 1:
		局_DB.Where("UPAgentId = ?", 0)
	case 2:
		局_DB.Where("UPAgentId != 0")
	}
	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1:
			局_DB.Where("Id = ?", 请求.Keywords)
		case 2:
			局_文本数组 := utils.Z正则_取全部匹配子文本(请求.Keywords, "([A-Za-z0-9]+)")
			if len(局_文本数组) == 1 {
				局_DB.Where("User  LIKE ?", "%"+请求.Keywords+"%")
			} else {
				局_DB.Where("User IN ? ", 局_文本数组)
			}
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

	var DB_User_简化实例 []DB_User_简化
	err := 局_DB.Count(&总数).Limit(请求.Size).Omit("login_app_name").Offset((请求.Page - 1) * 请求.Size).Find(&DB_User_简化实例).Error
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		global.GVA_LOG.Error("GetUserList:" + err.Error())
		return
	}

	var AppName = Ser_AppInfo.App取map列表String(true)
	for 索引 := range DB_User_简化实例 {
		DB_User_简化实例[索引].LoginAppName = AppName[DB_User_简化实例[索引].LoginAppid]
	}

	response.OkWithDetailed(struct {
		List  interface{} `json:"list"`
		Count int64       `json:"count"`
	}{DB_User_简化实例, 总数}, "获取成功", c)
}

// NewUser 新建用户
func (C *UserCtrl) NewUser(c *gin.Context) {
	var 请求 DB.DB_User
	if !C.ToJSON(c, &请求) {
		return
	}
	if 请求.Id != 0 {
		response.FailWithMessage("添加用户不能有id值", c)
		return
	}
	if 请求.Rmb != 0 && c.GetInt("Uid") != 1 {
		response.FailWithMessage("非系统管理员只能创建余额=0的普通用户", c)
		return
	}
	_, err := Ser_User.New用户信息(请求.User, 请求.PassWord, 请求.SuperPassWord, 请求.Qq, 请求.Email, 请求.Phone, c.ClientIP(), 请求.Note, 0, 0, 请求.Rmb, "")
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("添加成功", c)
	if 请求.Rmb != 0 {
		go Ser_Log.Log_写余额日志(请求.User, c.ClientIP(), fmt.Sprintf("管理员(%v),新增用户携带余额:%v", c.GetInt("Uid"), 请求.Rmb), 请求.Rmb)
	}
}

// SaveUser 保存用户信息
func (C *UserCtrl) SaveUser(c *gin.Context) {
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
	if 请求.SuperPassWord != "" && !utils.Z正则_校验密码(请求.SuperPassWord, &msg) {
		response.FailWithMessage("超级密码"+msg, c)
		return
	}

	用户详情, ok := Ser_User.Id取详情(请求.Id)
	if !ok {
		response.FailWithMessage("用户不存在", c)
		return
	}
	if 用户详情.Rmb != 请求.Rmb && c.GetInt("Uid") != 1 {
		response.FailWithMessage("非系统管理员不能通过编辑改变用户余额", c)
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

// SetUserStatus 批量修改用户状态
func (C *UserCtrl) SetUserStatus(c *gin.Context) {
	var 请求 请求_UserSetStatus
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
		局_user数组 := make([]string, 0, len(请求.Id))
		for _, 值 := range 请求.Id {
			局_user数组 = append(局_user数组, Ser_User.Id取User(值))
		}
		_ = Ser_LinkUser.Set批量注销User数组(局_user数组, Ser_LinkUser.Z注销_管理员手动注销)
	} else {
		err = global.GVA_DB.Model(DB.DB_User{}).Where("Id IN ? ", 请求.Id).Update("Status", 1).Error
	}
	if err != nil {
		response.FailWithMessage("修改失败", c)
		global.GVA_LOG.Error("修改失败:" + err.Error())
		return
	}
	response.OkWithMessage("修改成功", c)
}

// DeleteUser 批量删除用户
func (C *UserCtrl) DeleteUser(c *gin.Context) {
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
	if agent.L_agent.Q取Id数组中代理数量(c, 请求.Id) > 0 {
		response.FailWithMessage("包含代理级别用户,代理请前往代理管理-代理账号删除", c)
		return
	}

	var 影响行数 int64
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var 局_UserId = 请求.Id
		影响行数 = tx.Model(DB.DB_User{}).Where("Id IN ? ", 局_UserId).Delete("").RowsAffected

		var 局数组_AppId []int
		err2 := tx.Model(DB.DB_AppInfo{}).Select("AppId").Where("AppType IN ? ", []int{1, 2}).Scan(&局数组_AppId).Error
		if err2 != nil {
			return err2
		}
		for _, 局_appid := range 局数组_AppId {
			err2 = tx.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(局_appid)).Where("Uid IN ? ", 局_UserId).Delete("").Error
			if err2 != nil {
				return err2
			}
		}
		return err2
	})
	if err != nil {
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
}

// BatchAddRMB 批量增减余额
func (C *UserCtrl) BatchAddRMB(c *gin.Context) {
	var 请求 请求_UserBatchRMB
	if !C.ToJSON(c, &请求) {
		return
	}
	if len(请求.Id) == 0 {
		response.FailWithMessage("Id数组为空", c)
		return
	}
	if 请求.RMB == 0 {
		response.FailWithMessage("增减值不能为0", c)
		return
	}
	if c.GetInt("Uid") != 1 {
		response.FailWithMessage("只有管理员Admin可批量操作余额", c)
		return
	}

	err := Ser_User.Id余额增减_批量(请求.Id, utils.Float64取绝对值(请求.RMB), 请求.RMB > 0)
	if err != nil {
		response.FailWithMessage("操作失败:"+err.Error(), c)
		return
	}

	局_前缀 := "管理员批量增加余额,原因:"
	if 请求.RMB < 0 {
		局_前缀 = "管理员批量减少余额,原因:"
	}
	for _, 局_id := range 请求.Id {
		Ser_Log.Log_写余额日志(Ser_User.Id取User(局_id), c.ClientIP(), 局_前缀+请求.Note, 请求.RMB)
	}
	response.OkWithMessage("修改成功", c)
}
