package controller

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_Chare"
	"server/Service/Ser_Ka"
	"server/Service/Ser_KaClass"
	"server/Service/Ser_Log"
	"server/Service/Ser_UserClass"
	"server/Service/Ser_UserConfig"
	"server/global"
	"server/new/app/logic/common/agent"
	"server/new/app/logic/common/agentLevel"
	"server/new/app/logic/common/ka"
	"server/new/app/logic/common/kaClassUpPrice"
	"server/new/app/service"
	"server/structs/Http/response"
	DB "server/structs/db"
	"sort"
	"strconv"
)

type AgentKa struct{}

func NewAgentKaController() *AgentKa {
	return &AgentKa{}
}

type AgentKa单Id请求 struct {
	Id int `json:"id"`
}

type AgentKa列表请求 struct {
	AppId        int      `json:"appId"`
	Page         int      `json:"page"`
	Status       int      `json:"status"`
	RegisterTime []string `json:"registerTime"`
	UseTime      []string `json:"useTime"`
	KaClassId    int      `json:"kaClassId"`
	Num          int      `json:"num"`
	Size         int      `json:"size"`
	Type         int      `json:"type"`
	Keywords     string   `json:"keywords"`
	Order        int      `json:"order"`
	Child        int      `json:"child"`
}

type AgentKa类价格 struct {
	KaClassName string  `json:"kaClassName"`
	AgentMoney  float64 `json:"agentMoney"`
}

type AgentKa列表响应 struct {
	List      interface{}        `json:"list"`
	Count     int64              `json:"count"`
	AppType   int                `json:"appType"`
	UserClass map[int]string     `json:"userClass"`
	KaClass   map[int]AgentKa类价格 `json:"kaClass"`
	User      string             `json:"user"`
}

type AgentKa批量请求 struct {
	Id     []int `json:"id"`
	Status int   `json:"status"`
	AppId  int   `json:"appId"`
}

type AgentKa新增请求 struct {
	Id        int      `json:"id"`
	Number    int      `json:"number"`
	AdminNote string   `json:"adminNote"`
	KaName    []string `json:"kaName"`
}

type AgentKa精简 struct {
	Id            int     `json:"id" gorm:"column:Id;primarykey"`
	Name          string  `json:"name" gorm:"column:Name;comment:卡号"`
	VipTime       int64   `json:"vipTime" gorm:"column:VipTime;comment:增减时间秒数或点数"`
	RMb           float64 `json:"rMb" gorm:"column:RMb;type:decimal(10,2);default:0;comment:余额增减"`
	VipNumber     float64 `json:"vipNumber" gorm:"column:VipNumber;type:decimal(10,2);default:0;comment:积分增减"`
	UserClassId   int     `json:"userClassId" gorm:"column:UserClassId;comment:用户分类id"`
	UserClassName string  `json:"userClassName"`
	Num           int     `json:"num" gorm:"column:Num;comment:可以充值次数"`
	MaxOnline     int     `json:"maxOnline" gorm:"column:MaxOnline;comment:最大在线数"`
	RegisterTime  int64   `json:"registerTime"`
}

type Agent库存制卡请求 struct {
	Id        int    `json:"id"`
	Number    int    `json:"number"`
	AgentNote string `json:"agentNote"`
}

type AgentKa备注请求 struct {
	Id   []int  `json:"id"`
	Note string `json:"note"`
}

type AgentApp列表键值对 struct {
	AppId   int    `json:"appId"`
	AppName string `json:"appName"`
}

type AgentKa充值请求 struct {
	Ka   string `json:"ka"`
	User string `json:"user"`
}

type AgentKa模板请求 struct {
	AppId      int    `json:"appId"`
	KaTemplate string `json:"kaTemplate"`
}

func (A *AgentKa) GetInfo(c *gin.Context) {
	var 请求 AgentKa单Id请求
	if err := c.ShouldBindJSON(&请求); err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	var 局_卡号 DB.DB_Ka
	if err := global.GVA_DB.Model(DB.DB_Ka{}).Omit("AdminNote").Where("Id = ?", 请求.Id).First(&局_卡号).Error; err != nil {
		response.FailWithMessage("查询详细信息失败", c)
		return
	}

	局_接口, ok := c.Get("局_在线信息")
	if !ok {
		response.FailWithMessage("读取缓存在线信息失败", c)
		return
	}
	局_在线信息 := 局_接口.(DB.DB_LinksToken)

	局_制卡人数组 := agent.L_agent.Q取下级代理数组_user(c, []int{c.GetInt("Uid")})
	局_制卡人数组 = append(局_制卡人数组, 局_在线信息.User)
	if !S数组_是否存在(局_制卡人数组, 局_卡号.RegisterUser) {
		response.FailWithMessage("权限不足,只能读取自己制卡信息", c)
		return
	}
	response.OkWithDetailed(局_卡号, "获取成功", c)
}

func (A *AgentKa) GetKaList(c *gin.Context) {
	var 请求 AgentKa列表请求
	if err := c.ShouldBindJSON(&请求); err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	局_临时通用, _ := c.Get("局_在线信息")
	局_在线信息 := 局_临时通用.(DB.DB_LinksToken)
	局_制卡人数组 := []string{局_在线信息.User}
	if 请求.Child == 1 {
		局_制卡人数组 = append(agent.L_agent.Q取下级代理数组_user(c, []int{c.GetInt("Uid")}), 局_在线信息.User)
	}

	局_DB := global.GVA_DB.Model(DB.DB_Ka{}).Where("RegisterUser IN ?", 局_制卡人数组)
	if 请求.AppId != 0 {
		局_DB.Where("AppId = ?", 请求.AppId)
	}
	if 请求.Order == 1 {
		局_DB.Order("Id ASC")
	} else {
		局_DB.Order("Id DESC")
	}
	if 请求.Status == 1 || 请求.Status == 2 {
		局_DB.Where("Status = ?", 请求.Status)
	}
	if 请求.Num == 1 || 请求.Num == 2 {
		switch 请求.Num {
		case 1:
			局_DB.Where("Num = NumMax")
		case 2:
			局_DB.Where("Num < NumMax")
		}
	}
	if len(请求.RegisterTime) == 2 && 请求.RegisterTime[0] != "" && 请求.RegisterTime[1] != "" {
		局_开始, _ := strconv.ParseInt(请求.RegisterTime[0], 10, 64)
		局_结束, _ := strconv.ParseInt(请求.RegisterTime[1], 10, 64)
		局_DB.Where("RegisterTime > ?", 局_开始).Where("RegisterTime < ?", 局_结束+86400)
	}
	if len(请求.UseTime) == 2 && 请求.UseTime[0] != "" && 请求.UseTime[1] != "" {
		局_开始, _ := strconv.ParseInt(请求.UseTime[0], 10, 64)
		局_结束, _ := strconv.ParseInt(请求.UseTime[1], 10, 64)
		局_DB.Where("UseTime > ?", 局_开始).Where("UseTime < ?", 局_结束+86400)
	}
	if 请求.KaClassId != 0 {
		局_DB.Where("KaClassId = ?", 请求.KaClassId)
	}
	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1:
			局_DB.Where("Id = ?", 请求.Keywords)
		case 2:
			局_DB.Where("LOCATE(?, Name)>0 ", 请求.Keywords)
		case 4:
			局_DB.Where("LOCATE(?, AgentNote)>0 ", 请求.Keywords)
		case 5:
			局_DB.Where("RegisterUser=? ", 请求.Keywords)
		case 6:
			局_DB.Where("LOCATE(?, User)>0 ", 请求.Keywords)
		case 7:
			局_DB.Where("LOCATE(?, InviteUser)>0 ", 请求.Keywords)
		}
	}

	var 局_总数 int64
	var 局_列表 []DB.DB_Ka
	if err := 局_DB.Count(&局_总数).Omit("AdminNote").Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&局_列表).Error; err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		global.GVA_LOG.Error("GetKaList:" + err.Error())
		return
	}

	局_AppType := Ser_AppInfo.App取AppType(请求.AppId)
	局_UserClass := Ser_UserClass.UserClass取map列表Int(请求.AppId)
	局_可制卡号ID, _ := agent.L_agent.Id取代理可制卡类和可用代理功能列表(c, c.GetInt("Uid"))
	局_卡类信息数组, _ := Ser_KaClass.Id取详细信息_数组(局_可制卡号ID)
	局_卡类Map := make(map[int]AgentKa类价格, len(局_卡类信息数组))
	局_DBTx := *global.GVA_DB
	局_代理信息, _ := service.NewUser(c, &局_DBTx).Info(c.GetInt("Uid"))
	for 局_索引 := range 局_卡类信息数组 {
		if 局_代理信息.UPAgentId > 0 {
			局_代理调价, _, err := kaClassUpPrice.L_kaClassUpPrice.J计算代理调价(c, 局_卡类信息数组[局_索引].Id, 局_代理信息.UPAgentId)
			if err == nil && 局_代理调价 > 0 {
				局_卡类信息数组[局_索引].AgentMoney = Float64加float64(局_卡类信息数组[局_索引].AgentMoney, 局_代理调价, 2)
			}
		}
		if 请求.AppId == 局_卡类信息数组[局_索引].AppId {
			局_卡类Map[局_卡类信息数组[局_索引].Id] = AgentKa类价格{
				KaClassName: 局_卡类信息数组[局_索引].Name,
				AgentMoney:  局_卡类信息数组[局_索引].AgentMoney,
			}
		}
	}

	response.OkWithDetailed(AgentKa列表响应{
		List:      局_列表,
		Count:     局_总数,
		AppType:   局_AppType,
		UserClass: 局_UserClass,
		KaClass:   局_卡类Map,
		User:      局_在线信息.User,
	}, "获取成功", c)
}

func (A *AgentKa) Z追回卡号(c *gin.Context) {
	var 请求 struct {
		Id []int `json:"id"`
	}
	if err := c.ShouldBindJSON(&请求); err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if len(请求.Id) == 0 {
		response.FailWithMessage("Id数组为空", c)
		return
	}
	if len(请求.Id) != 1 {
		response.FailWithMessage("Id数组暂时只支持1个成员数,后续扩展中", c)
		return
	}
	if !Ser_Ka.Id检测制卡人(请求.Id, c.GetString("User")) {
		response.FailWithMessage("只能操作制卡人为本人的卡号", c)
		return
	}

	局_临时通用, _ := c.Get("局_在线信息")
	局_在线信息 := 局_临时通用.(DB.DB_LinksToken)
	if Ser_Ka.Id取制卡人(请求.Id[0]) != 局_在线信息.User {
		response.FailWithMessage("只能追回自己制造的卡号", c)
		return
	}
	if err := ka.L_ka.K卡号追回(c, 请求.Id[0], c.GetString("User")); err != nil {
		response.FailWithMessage("追回失败:"+err.Error(), c)
		return
	}

	局_卡号详情, _ := Ser_Ka.Id取详情(请求.Id[0])
	局_信息 := "操作卡号管理:代理追回卡号:" + 局_卡号详情.Name
	Ser_Log.Log_写代理操作日志(c.GetInt("Uid"), agentLevel.L_agentLevel.Q取Id代理级别(c, c.GetInt("Uid")), 局_卡号详情.AppId, 局_卡号详情.Id, 局_卡号详情.Name, DB.D代理功能_卡号追回, c.ClientIP(), 局_信息)
	response.OkWithMessage("操作成功", c)
}

func (A *AgentKa) New(c *gin.Context) {
	var 请求 AgentKa新增请求
	if err := c.ShouldBindJSON(&请求); err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if 请求.Number <= 0 {
		response.FailWithMessage("生成数量必须大于0", c)
		return
	}
	if 请求.Number > 5000 {
		response.FailWithMessage("生成数量每批最大5000", c)
		return
	}
	if !agent.L_agent.Id卡类权限检测(c, c.GetInt("Uid"), 请求.Id) {
		response.FailWithMessage("无该卡制卡权限", c)
		return
	}
	if !Ser_KaClass.KaClassId是否存在(请求.Id) {
		response.FailWithMessage("卡类id不存在", c)
		return
	}

	局_卡数组 := make([]DB.DB_Ka, 请求.Number)
	局_接口, ok := c.Get("局_在线信息")
	if !ok {
		response.FailWithMessage("读取缓存在线信息失败", c)
		return
	}
	局_卡类信息, err := Ser_KaClass.KaClass取详细信息(请求.Id)
	if err != nil {
		response.FailWithMessage("卡类id不存在", c)
		return
	}
	局_在线信息 := 局_接口.(DB.DB_LinksToken)
	if err = Ser_Ka.Ka代理批量购买(c, 局_卡数组[:], 请求.Id, 局_在线信息.Uid, 请求.AdminNote, 0, c.ClientIP()); err != nil {
		response.FailWithMessage("制卡失败:"+err.Error(), c)
		return
	}

	局_用户类型名称 := ""
	if 局_用户类型, ok := Ser_UserClass.Id取详情(局_卡类信息.AppId, 局_卡类信息.UserClassId); ok {
		局_用户类型名称 = 局_用户类型.Name
	}
	局_精简列表 := A.转精简卡列表(局_卡数组, 局_用户类型名称)
	response.OkWithDetailed(局_精简列表, "制卡成功", c)
}

func (A *AgentKa) K库存制卡(c *gin.Context) {
	var 请求 Agent库存制卡请求
	if err := c.ShouldBindJSON(&请求); err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}

	局_卡数组 := make([]DB.DB_Ka, 请求.Number)
	if err := Ser_Ka.Ka代理批量库存购买(c, 局_卡数组[:], 请求.Id, 请求.Number, c.GetInt("Uid"), 请求.AgentNote, c.ClientIP()); err != nil {
		response.FailWithMessage("制卡失败:"+err.Error(), c)
		return
	}

	局_用户类型名称 := ""
	if 局_用户类型, ok := Ser_UserClass.Id取详情(局_卡数组[0].AppId, 局_卡数组[0].UserClassId); ok {
		局_用户类型名称 = 局_用户类型.Name
	}
	局_精简列表 := A.转精简卡列表(局_卡数组, 局_用户类型名称)
	response.OkWithDetailed(局_精简列表, "制卡成功", c)
}

func (A *AgentKa) Set修改状态(c *gin.Context) {
	var 请求 AgentKa批量请求
	if err := c.ShouldBindJSON(&请求); err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if len(请求.Id) == 0 {
		response.FailWithMessage("Id数组为空", c)
		return
	}

	局_权限文本 := "卡号状态参数错误"
	局_代理权限ID := 0
	局_权限 := false
	switch 请求.Status {
	case 1:
		局_权限 = agent.L_agent.Id功能权限检测(c, c.GetInt("Uid"), DB.D代理功能_卡号解冻)
		局_权限文本 = "无卡号解冻权限,请联系上级代理授权"
		局_代理权限ID = DB.D代理功能_卡号解冻
	case 2:
		局_权限 = agent.L_agent.Id功能权限检测(c, c.GetInt("Uid"), DB.D代理功能_卡号冻结)
		局_权限文本 = "无卡号冻结权限,请联系上级代理授权"
		局_代理权限ID = DB.D代理功能_卡号冻结
	}
	if !局_权限 {
		response.FailWithMessage(局_权限文本, c)
		return
	}
	if !Ser_Ka.Id检测制卡人(请求.Id, c.GetString("User")) {
		response.FailWithMessage("只能操作制卡人为本人的卡号", c)
		return
	}

	if err := Ser_Ka.Ka修改状态_同步卡号模式软件用户(请求.Id, 请求.Status); err != nil {
		response.FailWithMessage("修改失败", c)
		global.GVA_LOG.Error("修改失败:" + err.Error())
		return
	}

	for _, 局_Id := range 请求.Id {
		局_卡号, err := Ser_Ka.Id取详情(局_Id)
		if err == nil {
			局_信息 := "操作卡号管理:"
			if 局_代理权限ID == DB.D代理功能_卡号冻结 {
				局_信息 += "卡号冻结"
			} else {
				局_信息 += "卡号解冻"
			}
			Ser_Log.Log_写代理操作日志(c.GetInt("Uid"), agentLevel.L_agentLevel.Q取Id代理级别(c, c.GetInt("Uid")), 局_卡号.AppId, 局_卡号.Id, 局_卡号.Name, 局_代理权限ID, c.ClientIP(), 局_信息)
		}
	}
	response.OkWithMessage("修改成功", c)
}

func (A *AgentKa) G更换卡号(c *gin.Context) {
	var 请求 AgentKa单Id请求
	if err := c.ShouldBindJSON(&请求); err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if !Ser_Ka.Id检测制卡人([]int{请求.Id}, c.GetString("User")) {
		response.FailWithMessage("只能操作制卡人为本人的卡号", c)
		return
	}

	局_旧卡号详情, _ := Ser_Ka.Id取详情(请求.Id)
	if err := Ser_Ka.Ka更换卡号(c, 请求.Id, c.GetInt("Uid"), c.ClientIP()); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	局_卡号详情, _ := Ser_Ka.Id取详情(请求.Id)
	局_信息 := "操作卡号管理:卡号更换新卡号:" + 局_卡号详情.Name
	Ser_Log.Log_写代理操作日志(c.GetInt("Uid"), agentLevel.L_agentLevel.Q取Id代理级别(c, c.GetInt("Uid")), 局_卡号详情.AppId, 局_卡号详情.Id, 局_旧卡号详情.Name, DB.D代理功能_更换卡号, c.ClientIP(), 局_信息)
	response.OkWithDetailed(局_卡号详情, "更换成功", c)
}

func (A *AgentKa) Set修改代理备注(c *gin.Context) {
	var 请求 AgentKa备注请求
	if err := c.ShouldBindJSON(&请求); err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if len(请求.Id) == 0 {
		response.FailWithMessage("Id数组为空", c)
		return
	}
	if !Ser_Ka.Id检测制卡人(请求.Id, c.GetString("User")) {
		response.FailWithMessage("只能操作制卡人为本人的卡号", c)
		return
	}

	局_接口, ok := c.Get("局_在线信息")
	if !ok {
		response.FailWithMessage("读取缓存在线信息失败", c)
		return
	}
	局_在线信息 := 局_接口.(DB.DB_LinksToken)
	if err := Ser_Ka.Ka修改代理备注(局_在线信息.User, 请求.Id, 请求.Note); err != nil {
		response.FailWithMessage("修改失败", c)
		global.GVA_LOG.Error("修改失败:" + err.Error())
		return
	}
	response.OkWithMessage("修改成功", c)
}

func (A *AgentKa) GetAppIdNameList(c *gin.Context) {
	局_AppIdName := Ser_AppInfo.App取map列表String(true)
	局_可操作应用Id := agent.L_agent.Id取代理可操作应用AppId列表(c, c.GetInt("Uid"))
	局_数组 := make([]AgentApp列表键值对, 0, len(局_AppIdName))
	for 局_索引 := range 局_可操作应用Id {
		局_数组 = append(局_数组, AgentApp列表键值对{
			AppId:   局_可操作应用Id[局_索引],
			AppName: 局_AppIdName[strconv.Itoa(局_可操作应用Id[局_索引])],
		})
	}
	for 局_AppId := range 局_AppIdName {
		if !S数组_整数是否存在(局_可操作应用Id, D到整数(局_AppId)) {
			delete(局_AppIdName, 局_AppId)
		}
	}
	sort.Slice(局_数组, func(i, j int) bool {
		return 局_数组[i].AppId < 局_数组[j].AppId
	})
	局_响应数组 := make([]Agent应用键值对, 0, len(局_数组))
	for 局_索引 := range 局_数组 {
		局_响应数组 = append(局_响应数组, Agent应用键值对{
			AppId:   局_数组[局_索引].AppId,
			AppName: 局_数组[局_索引].AppName,
		})
	}
	response.OkWithDetailed(Agent应用列表响应{Map: 局_AppIdName, Array: 局_响应数组}, "获取成功", c)
}

func (A *AgentKa) K卡号充值(c *gin.Context) {
	var 请求 AgentKa充值请求
	if err := c.ShouldBindJSON(&请求); err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}
	if err := ka.L_ka.K卡号充值_事务(c, 0, 请求.Ka, 请求.User, ""); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("充值成功", c)
}

func (A *AgentKa) Get卡号列表统计制卡(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get卡号列表统计制卡_代理(c), "获取成功", c)
}

func (A *AgentKa) Set修改卡号生成模板(c *gin.Context) {
	var 请求 AgentKa模板请求
	if err := c.ShouldBindJSON(&请求); err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if err := Ser_UserConfig.Z置值(1, c.GetInt("Uid"), "卡号生成格式模板"+strconv.Itoa(请求.AppId), 请求.KaTemplate); err != nil {
		response.FailWithMessage("修改失败", c)
		global.GVA_LOG.Error("修改失败:" + err.Error())
		return
	}
	response.OkWithMessage("修改成功", c)
}

func (A *AgentKa) Q取卡号生成模板(c *gin.Context) {
	var 请求 AgentKa模板请求
	if err := c.ShouldBindJSON(&请求); err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	局_模板 := Ser_UserConfig.Q取值(1, c.GetInt("Uid"), "卡号生成格式模板"+strconv.Itoa(请求.AppId))
	if 局_模板 == "" {
		局_模板 = "卡号:{Name} "
		if Ser_AppInfo.App是否为计点(请求.AppId) {
			局_模板 += "点数"
		} else {
			局_模板 += "时间"
		}
		局_模板 += ":{VipTime} 积分:{VipTime} 软件:{AppName} 余额:{RMb} 积分:{VipNumber}"
	}
	response.OkWithData(局_模板, c)
}

func (A *AgentKa) 转精简卡列表(卡数组 []DB.DB_Ka, 用户类型名称 string) []AgentKa精简 {
	局_结果 := make([]AgentKa精简, len(卡数组))
	for 局_索引 := range 局_结果 {
		局_结果[局_索引].Name = 卡数组[局_索引].Name
		局_结果[局_索引].Id = 卡数组[局_索引].Id
		局_结果[局_索引].RMb = 卡数组[局_索引].RMb
		局_结果[局_索引].VipTime = 卡数组[局_索引].VipTime
		局_结果[局_索引].VipNumber = 卡数组[局_索引].VipNumber
		局_结果[局_索引].UserClassId = 卡数组[局_索引].UserClassId
		局_结果[局_索引].UserClassName = 用户类型名称
		局_结果[局_索引].Num = 卡数组[局_索引].Num
		局_结果[局_索引].MaxOnline = 卡数组[局_索引].MaxOnline
		局_结果[局_索引].RegisterTime = 卡数组[局_索引].RegisterTime
	}
	return 局_结果
}
