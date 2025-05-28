package AgentInventory

import (
	"EFunc/utils"
	"github.com/gin-gonic/gin"
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
// 获取用户信息列表
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
	局_DB := global.GVA_DB.Model(DB.Db_Agent_库存卡包{})
	局_DB2 := global.GVA_DB.Model(DB.Db_Agent_库存卡包{})

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
func (a *Api) New库存包信息(c *gin.Context) {
	var 请求 DB.Db_Agent_库存卡包
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}

	if 请求.Id != 0 {
		response.FailWithMessage("添加库存卡包不能有id值", c)
		return
	}

	if !utils.S数组_整数是否存在(agent.L_agent.Q取下级代理数组含子级(c, []int{-c.GetInt("Uid")}), 请求.Uid) {
		response.FailWithMessage("只能给自己的下级或子级代理创建库存", c)
		return
	}

	新库存卡包, err2 := Ser_AgentInventory.New(c, 请求.Uid, 请求.KaClassId, 请求.NumMax, -c.GetInt("Uid"), -1, 请求.EndTime, 请求.Note)
	if err2 != nil {
		response.FailWithMessage(err2.Error(), c)
		return
	}
	response.OkWithMessage("操作成功", c)

	User1角色 := agentLevel.L_agentLevel.Q取Id代理级别(c, 请求.Uid)
	if User1角色 == 0 {
		User1角色 = 4
	}
	Ser_Log.Log_写库存转移日志(新库存卡包.Id, 新库存卡包.NumMax, 2, agent.L_agent.ID取用户名(c, 请求.Uid), User1角色, agent.L_agent.ID取用户名(c, -c.GetInt("Uid")), 4, c.ClientIP(), "接收管理员库存包")

	return
}

func (a *Api) Del批量删除库存(c *gin.Context) {
	var 请求 结构请求_ID数组
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

	var 影响行数 int64
	var db = global.GVA_DB

	影响行数 = db.Model(DB.Db_Agent_库存卡包{}).Where("Id IN ? ", 请求.Id).Delete("").RowsAffected
	if db.Error != nil {
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
	return
}

type 结构请求_ID数组 struct {
	Id []int `json:"Id"` //用户id数组
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
	//管理员Uid为负数
	err = Ser_AgentInventory.K库存撤回(c, -c.GetInt("Uid"), 请求.Id, 请求.Num, 请求.Note, c.ClientIP())
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("撤回成功", c)
	return
}

// 获取代理可制卡类授权
func (a *Api) Get取下级代理列表和可创建库存包列表(c *gin.Context) {
	var 请求 结构请求_单id
	//{"Id":2} 代理用户ID
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}
	if 请求.Id == 0 {
		response.FailWithMessage("代理ID不能为0", c)
		return
	}
	var 局_代理信息 = []DB.Db_Agent_Level{}
	global.GVA_DB.Model(DB.Db_Agent_Level{}).Where("UPAgentId = ?", 请求.Id).Find(&局_代理信息)
	var 数组_Uid = make([]int, len(局_代理信息))
	for 索引, 值 := range 局_代理信息 {
		数组_Uid[索引] = 值.Uid
	}

	//局_耗时 := time.Now().Unix()
	var 局_用户数组 []DB.DB_User

	_ = global.GVA_DB.Model(DB.DB_User{}).Select("Id", "User", "UPAgentId", "AgentDiscount").Where("Id In ?", 数组_Uid).Find(&局_用户数组).Error
	nodes := make([]*Node, 0, len(局_用户数组))
	for 索引, _ := range 局_用户数组 {
		nodes = append(nodes, &Node{
			Id:            局_用户数组[索引].Id,
			UPAgentId:     局_用户数组[索引].UPAgentId,
			User:          局_用户数组[索引].User,
			AgentDiscount: 局_用户数组[索引].AgentDiscount,
		})
	}
	var 响应数据 结构请求_代理树和卡类树
	var 局_uid int
	局_uid = -c.GetInt("Uid")
	响应数据.AgentTree = 转换为代理树2(nodes, 局_uid) //只能给自己的上级代理添加库存
	局_临时 := Ser_KaClass.Q取全部可制卡类树形框列表(c, 请求.Id)
	响应数据.KaClassTree = 局_临时
	response.OkWithDetailed(响应数据, "获取成功", c)
	return
}

type 结构请求_代理树和卡类树 struct {
	AgentTree []*Node `json:"AgentTree"`

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
	if 代理列表 == nil {
		return []*Node{}
	}

	memo := make(map[int]*Node)
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
			// 确保 memo[v.UPAgentId] 不为 nil
			if memo[v.UPAgentId] == nil {
				memo[v.UPAgentId] = &Node{Children: []*Node{}}
			}
			memo[v.UPAgentId] = &Node{Children: []*Node{memo[v.Id]}}
		}
	}
	// 安全返回
	if parent, ok := memo[上级ID]; ok {
		return parent.Children
	}
	return []*Node{}

}
