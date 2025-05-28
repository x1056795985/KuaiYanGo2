package AgentInventory

import (
	"github.com/gin-gonic/gin"
	"server/Service/Ser_Admin"
	"server/Service/Ser_AgentInventory"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_KaClass"
	"server/Service/Ser_Log"
	"server/Service/Ser_User"
	"server/global"
	"server/new/app/logic/common/agent"
	"server/new/app/logic/common/agentLevel"
	"server/structs/Http/response"
	DB "server/structs/db"
	"strconv"
	"time"
)

type Api struct{}

// GetInfo
func (a *Api) GetAgentInventoryInfo(c *gin.Context) {
	var 请求 结构请求_单id
	//{"Id":2}
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	if Ser_AgentInventory.Id取归属Uid(请求.Id) != c.GetInt("Uid") {
		response.FailWithMessage("只能查看自己的库存详细信息", c)
		return
	}

	var 卡号库存卡包 DB.Db_Agent_库存卡包

	err = global.GVA_DB.Model(DB.Db_Agent_库存卡包{}).Where("id = ?", 请求.Id).First(&卡号库存卡包).Error
	// 没查到数据
	if err != nil {
		response.FailWithMessage("查询详细信息失败", c)
		return
	}

	response.OkWithDetailed(卡号库存卡包, "获取成功", c)
	return
}

type DB_AgentInventory2 struct {
	DB.DB_User
	LoginAppName     string `json:"LoginAppName"`     //登录平台App名字
	Role             int    `json:"Role"`             //
	UPAgentInventory string `json:"UPAgentInventory"` //
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

// GetList
// 获取列表
func (a *Api) GetAgentInventoryList(c *gin.Context) {
	var 请求 结构请求_GetUserList
	//{"Type":"2","Size":10,"Page":1,"Status":"1","keywords":"1"}
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	var 总数 int64
	//筛选归属为我的库存包,和我转出的库存包  不可  局_DB2:=局_DB方式, 因为复制的是指针,并不是真完全复制,后面执行会出错误
	局_DB := global.GVA_DB.Model(DB.Db_Agent_库存卡包{}).
		Where("(Uid = ? OR RegisterUserId = ? OR SourceUid=? )", c.GetInt("Uid"), c.GetInt("Uid"), c.GetInt("Uid"))
	局_DB2 := global.GVA_DB.Model(DB.Db_Agent_库存卡包{}).
		Where("(Uid = ? OR RegisterUserId = ? OR SourceUid=? )", c.GetInt("Uid"), c.GetInt("Uid"), c.GetInt("Uid"))

	if 请求.Order == 1 {
		局_DB.Order("Id ASC")
		局_DB2.Order("Id ASC")
	} else {
		局_DB.Order("Id DESC")
		局_DB2.Order("Id DESC")
	}

	/*            <el-option key="1" label="正常" :value="1"/>
	<el-option key="2" label="过期" :value="2"/>
	<el-option key="3" label="耗尽" :value="3"/>*/
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
		case 1: //id
			局_DB.Where("Id = ?", 请求.Keywords)
			局_DB2.Where("Id = ?", 请求.Keywords)
		case 2: //用户名
			局_id := Ser_User.User用户名取id(请求.Keywords)
			局_DB.Where("Uid= ? ", 局_id)
			局_DB2.Where("Uid= ? ", 局_id)
		case 3: //库存剩余>?
			atoi, err2 := strconv.Atoi(请求.Keywords)
			if err2 != nil {
				response.FailWithMessage("库存剩余数量只能为整数", c)
				return
			}
			局_DB.Where("(NumMax-Num) > ?", atoi)
			局_DB2.Where("(ai.NumMax-ai.Num) > ?", atoi)
		case 4: //备注
			局_DB.Where("LOCATE(?, Note)>0 ", 请求.Keywords)
			局_DB2.Where("LOCATE(?, ai.Note)>0 ", 请求.Keywords)
		}
	}
	var 数组_库存卡包 []Db_Agent_库存卡包_扩展1
	//Count(&总数) 必须放在where 后面 不然值会被清0
	局_DB.Count(&总数)
	err = 局_DB2.Table("db_Agent_Inventory ai").
		Select("ai.*, u.User, kc.Name AS KaClassName,kc.AppId").
		Joins("LEFT JOIN db_User u ON ai.Uid = u.Id").
		Joins("LEFT JOIN db_Ka_Class kc ON ai.KaClassId = kc.Id").
		Omit("AppName").
		Limit(请求.Size).
		Offset((请求.Page - 1) * 请求.Size).
		Find(&数组_库存卡包).Error

	局_map := Ser_AppInfo.AppInfo取map列表Int()
	for 索引, _ := range 数组_库存卡包 {
		数组_库存卡包[索引].AppName = 局_map[数组_库存卡包[索引].AppId]
	}

	response.OkWithDetailed(结构响应_GetUserList{数组_库存卡包, 总数}, "获取成功", c)
	return

}

type Db_Agent_库存卡包_扩展1 struct {
	DB.Db_Agent_库存卡包
	User        string `json:"User" gorm:"column:User;index;comment:用户名"`
	KaClassName string `json:"KaClassName" gorm:"column:KaClassName;index;comment:卡类名称"`
	AppId       int    `json:"AppId" gorm:"column:AppId;应用Id"`
	AppName     string `json:"AppName" gorm:"column:AppName;应用名称"`
}
type 结构响应_GetUserList struct {
	List  interface{} `json:"List"`  // 列表
	Count int64       `json:"Count"` // 总数
}

// New
func (a *Api) New库存购买(c *gin.Context) {
	var 请求 DB.Db_Agent_库存卡包
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	请求.Id = 0
	请求.Uid = c.GetInt("Uid")
	新库存卡包, err2 := Ser_AgentInventory.New代理购买(c, 请求.Uid, 请求.KaClassId, 请求.NumMax, 请求.EndTime, 请求.Note, c.ClientIP())
	if err2 != nil {
		response.FailWithMessage(err2.Error(), c)
		return
	}

	response.OkWithMessage("操作成功", c)

	User1角色 := agentLevel.L_agentLevel.Q取Id代理级别(c, c.GetInt("Uid"))
	if User1角色 == 0 {
		User1角色 = 4
	}
	var 局_创建用户名 = ""
	if 请求.Uid < 0 {
		局_创建用户名 = Ser_Admin.Id取User(请求.Uid)
	} else {
		局_创建用户名 = agent.L_agent.ID取用户名(c, 请求.Uid)
	}

	Ser_Log.Log_写库存转移日志(新库存卡包.Id, 新库存卡包.NumMax, 3, 局_创建用户名, User1角色, 局_创建用户名, User1角色, c.ClientIP(), "自助购买")
	return
}

type 结构请求_库存撤回 struct {
	Id   int    `json:"Id"`   //库存Id
	Num  int    `json:"Num"`  //撤回数量
	Note string `json:"Note"` //备注
}

func (a *Api) K库存撤回(c *gin.Context) {
	var 请求 结构请求_库存撤回
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}

	err = Ser_AgentInventory.K库存撤回(c, c.GetInt("Uid"), 请求.Id, 请求.Num, 请求.Note, c.ClientIP())
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("撤回成功", c)
	return
}

type 结构请求_库存发送 struct {
	SourceID int    `json:"SourceID"` //原库存Id
	Num      int    `json:"Num"`      //发送数量
	ToUserId int    `json:"ToUserId"` //目标代理Id
	Note     string `json:"Note"`     //备注

}

func (a *Api) K库存发送(c *gin.Context) {
	var 请求 结构请求_库存发送
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if Ser_AgentInventory.Id取归属Uid(请求.SourceID) != c.GetInt("Uid") {
		response.FailWithMessage("只能将归属自己的库存,发送给别人.", c)
		return
	}

	err = Ser_AgentInventory.K库存发送(c, 请求.SourceID, 请求.ToUserId, 请求.Num, 请求.Note, c.ClientIP())
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("发送成功", c)
	return
}
func (a *Api) K库存延期(c *gin.Context) {
	var 请求 结构请求_库存撤回
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}

	err = Ser_AgentInventory.K库存延期(请求.Id, c.GetInt("Uid"), 请求.Num)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("操作成功", c)
	return
}
func (a *Api) K库存修改备注(c *gin.Context) {
	var 请求 结构请求_库存撤回
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}

	err = Ser_AgentInventory.K库存修改备注(请求.Id, c.GetInt("Uid"), 请求.Note)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("操作成功", c)
	return
}

// 获取代理可制卡类授权
func (a *Api) Get取可创建库存包列表(c *gin.Context) {
	var 请求 结构请求_单id
	//{"Id":2} 代理用户ID
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}
	局_临时 := Ser_KaClass.Q取全部可制卡类树形框列表(c, c.GetInt("Uid"))
	response.OkWithDetailed(结构请求_代理树和卡类树{局_临时}, "获取成功", c)
	return
}

type 结构请求_代理树和卡类树 struct {
	KaClassTree []Ser_KaClass.K可制卡类授权树形框结构 `json:"KaClassTree"`
}

type Node struct {
	Id            int     `json:"Id" gorm:"column:Id;primarykey;AUTO_INCREMENT"` // id
	User          string  `json:"User" gorm:"column:User;size:191;UNIQUE;index;comment:用户登录名"`
	UPAgentId     int     `json:"UPAgentId" gorm:"column:UPAgentId;comment:上级代理id"`
	AgentDiscount int     `json:"AgentDiscount" gorm:"column:AgentDiscount;comment:分成百分比"`
	Children      []*Node `json:"Children,omitempty" gorm:"column:Children;comment:下级代理id"`
}

func 转换为代理树2(代理列表 []*Node, 上级ID int) []*Node {

	memo := make(map[int]*Node)
	if 代理列表 == nil {
		return memo[上级ID].Children
	}

	for _, v := range 代理列表 {
		if _, ok := memo[v.Id]; ok {
			v.Children = memo[v.Id].Children
			memo[v.Id] = v
		} else {
			v.Children = make([]*Node, 0)
			memo[v.Id] = v
		}
		if _, ok := memo[v.UPAgentId]; ok {
			memo[v.UPAgentId].Children = append(memo[v.UPAgentId].Children, memo[v.Id])
		} else {
			memo[v.UPAgentId] = &Node{Children: []*Node{memo[v.Id]}}
		}
	}
	return memo[上级ID].Children

}

func (a *Api) Q可发送库存下级代理(c *gin.Context) {

	局_数组_下级代理ID := agent.L_agent.Q取下级代理数组(c, []int{c.GetInt("Uid")})
	if len(局_数组_下级代理ID) == 0 {
		response.FailWithMessage("无直属下级代理", c)
		return
	}
	局_数组_下级代理详细信息, err := Ser_User.Id取详情_数组(局_数组_下级代理ID)
	if err != nil {
		response.FailWithMessage("读取失败:"+err.Error(), c)
		return
	}

	局_可发送下级列表 := make([]下级代理, 0, len(局_数组_下级代理详细信息))

	for 索引 := range 局_数组_下级代理详细信息 {
		局_可发送下级列表 = append(局_可发送下级列表, 下级代理{
			Id:       局_数组_下级代理详细信息[索引].Id,
			User:     局_数组_下级代理详细信息[索引].User,
			Disabled: 局_数组_下级代理详细信息[索引].Status == 2, //如果值是冻结则为真
		})
	}

	response.OkWithDetailed(局_可发送下级列表, "成功", c)
	return

}

type 下级代理 struct {
	Id       int    `json:"Id"`
	User     string `json:"User"`     //名称
	Disabled bool   `json:"disabled"` //禁用
}
