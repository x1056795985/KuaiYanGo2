package Ka

import (
	"github.com/gin-gonic/gin"
	"server/Service/Ser_Agent"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_Chare"
	"server/Service/Ser_Ka"
	"server/Service/Ser_KaClass"
	"server/Service/Ser_UserClass"
	"server/global"
	"server/structs/Http/response"
	DB "server/structs/db"
	"strconv"
)

type Api struct{}

// GetKaInfo
func (a *Api) GetInfo(c *gin.Context) {
	var 请求 结构请求_单id
	//{"Id":2}
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	var DB_Ka DB.DB_Ka

	err = global.GVA_DB.Model(DB.DB_Ka{}).Omit("AdminNote").Where("Id = ?", 请求.Id).First(&DB_Ka).Error
	// 没查到数据
	if err != nil {
		response.FailWithMessage("查询详细信息失败", c)
		return
	}
	局_接口, ok := c.Get("局_在线信息")
	if !ok {
		response.FailWithMessage("读取缓存在线信息失败", c)
		return
	}
	局_在线信息 := 局_接口.(DB.DB_LinksToken)
	if DB_Ka.RegisterUser != 局_在线信息.User {
		response.FailWithMessage("权限不足,只能读取自己制卡信息", c)
		return
	}

	response.OkWithDetailed(DB_Ka, "获取成功", c)
	return
}

type 结构请求_单id struct {
	Id int `json:"Id"`
}

type 结构请求_GetKaList struct {
	AppId        int      `json:"AppId"`        // Appid 必填
	Page         int      `json:"Page"`         // 页
	Status       int      `json:"Status"`       // 状态
	RegisterTime []string `json:"RegisterTime"` // 制卡开始时间 制卡结束时间
	KaClassId    int      `json:"KaClassId"`    // 卡类id
	Num          int      `json:"Num"`          // 卡使情况
	Size         int      `json:"Size"`         // 页数量
	Type         int      `json:"Type"`         // 关键字类型  1 id 2 用户名 3绑定信息 4 动态标签
	Keywords     string   `json:"Keywords"`     // 关键字
	Order        int      `json:"Order"`        // 0 倒序 1 正序
}

// GetKaList
// 获取用户信息列表
func (a *Api) GetKaList(c *gin.Context) {
	var 请求 结构请求_GetKaList
	//{"AppId":2,"Type":2,"Size":10,"Page":1,"Status":1,"keywords":"1"}
	err := c.ShouldBindJSON(&请求)

	//解析失败
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}
	局_临时通用, _ := c.Get("局_在线信息")
	局_在线信息 := 局_临时通用.(DB.DB_LinksToken)
	/*
		并不需要,直接限制制卡人值就可以了,
		局_AppID列表 := Ser_Agent.Id取代理可操作应用AppId列表(c.GetInt("Uid"))
			if 请求.AppId != 0 && utils.S数组_整数是否存在(局_AppID列表, 请求.AppId) {
				response.FailWithMessage("无该应用权限,请联系上级授权该应用对应卡类制卡权限", c)
				return
			}*/

	var DB_Ka []DB.DB_Ka
	var 总数 int64

	局_DB := global.GVA_DB.Model(DB.DB_Ka{}).Where("RegisterUser=?", 局_在线信息.User) //直接限制只允许读取制卡人为自己的卡号
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
		case 1: //已经使用
			局_DB.Where("Num = NumMax")
		case 2: //未使用过
			局_DB.Where("Num < NumMax")
		}
	}
	if 请求.RegisterTime != nil && len(请求.RegisterTime) == 2 && 请求.RegisterTime[0] != "" && 请求.RegisterTime[1] != "" {
		制卡开始时间, _ := strconv.Atoi(请求.RegisterTime[0])
		制卡结束时间, _ := strconv.Atoi(请求.RegisterTime[1])
		局_DB.Where("RegisterTime > ?", 制卡开始时间).Where("RegisterTime < ?", 制卡结束时间+86400)
	}

	if 请求.KaClassId != 0 {
		局_DB.Where("KaClassId = ?", 请求.KaClassId)
	}

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //id
			局_DB.Where("Id = ?", 请求.Keywords)
		case 2: //卡号
			局_DB.Where("LOCATE(?, Name)>0 ", 请求.Keywords)
		case 3: //管理员备注
		/*	局_DB.Where("LOCATE(?, AdminNote)>0 ", 请求.Keywords)*/
		case 4: //代理备注
			局_DB.Where("LOCATE(?, AgentNote)>0 ", 请求.Keywords)
		case 5: //制卡人
			局_DB.Where("RegisterUser=? ", 请求.Keywords)
		case 6: //充值用户
			局_DB.Where("LOCATE(?, User)>0 ", 请求.Keywords)
		case 7: //推荐人
			局_DB.Where("LOCATE(?, InviteUser)>0 ", 请求.Keywords)
		}
	}

	//Count(&总数) 必须放在where 后面 不然值会被清0  不让代理看管理员备注
	err = 局_DB.Count(&总数).Omit("AdminNote").Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&DB_Ka).Error
	//fmt.Println("用户总数%d", 总数, DB_LinksToken)
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		global.GVA_LOG.Error("GetKaList:" + err.Error())
		return
	}
	var AppType int
	AppType = Ser_AppInfo.App取AppType(请求.AppId)
	UserClass := Ser_UserClass.UserClass取map列表Int(请求.AppId)

	可制卡号ID, _ := Ser_Agent.Id取代理可制卡类和可用代理功能列表(c.GetInt("Uid"))
	局_卡类信息数组, _ := Ser_KaClass.Id取详细信息_数组(可制卡号ID)
	var KaClass2 = make(map[int]结构响应_卡类名称价格, len(局_卡类信息数组))
	for 索引 := range 局_卡类信息数组 {
		if 请求.AppId == 局_卡类信息数组[索引].AppId {
			KaClass2[局_卡类信息数组[索引].Id] = 结构响应_卡类名称价格{局_卡类信息数组[索引].Name, 局_卡类信息数组[索引].AgentMoney}
		}
	}

	response.OkWithDetailed(结构响应_GetKaList{DB_Ka, 总数, AppType, UserClass, KaClass2}, "获取成功", c)
	return
}

type 结构响应_GetKaList struct {
	List      interface{}         `json:"List"`      // 列表
	Count     int64               `json:"Count"`     // 总数
	AppType   int                 `json:"AppType"`   //
	UserClass map[int]string      `json:"UserClass"` //
	KaClass   map[int]结构响应_卡类名称价格 `json:"KaClass"`   //
}
type 结构响应_卡类名称价格 struct {
	KaClassName string  `json:"KaClassName"` //
	AgentMoney  float64 `json:"AgentMoney"`  //

}

// Z追回卡号 已用充值卡将相应的卡使用者和推荐人强行扣回充值卡面值,可能扣成负数
func (a *Api) Z追回卡号(c *gin.Context) {
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

	if len(请求.Id) != 1 {
		response.FailWithMessage("Id数组暂时只支持1个成员数,后续扩展中", c)
		return
	}
	if !Ser_Agent.Id功能权限检测(c.GetInt("Uid"), DB.D代理功能_卡号追回) {
		response.FailWithMessage("权限不足,无卡号追回权限,请联系上级授权", c)
		return
	}

	局_临时通用, _ := c.Get("局_在线信息")
	局_在线信息 := 局_临时通用.(DB.DB_LinksToken)

	if Ser_Ka.Id取制卡人(请求.Id[0]) != 局_在线信息.User {
		response.FailWithMessage("只能追回自己制造的卡号", c)
		return
	}

	提示, err := Ser_Ka.K卡号追回(请求.Id[0])

	if err != nil {
		response.FailWithMessage("追回失败:"+err.Error(), c)
		return
	}

	response.OkWithMessage(提示, c)
	return
}

type 结构请求_ID数组 struct {
	Id    []int `json:"Id"` //用户id数组
	AppId int   `json:"AppId"`
}

type 结构请求_New struct {
	Id        int      `json:"Id"`        //卡类id
	Number    int      `json:"Number"`    //生成数量
	AdminNote string   `json:"AdminNote"` //管理员备注
	KaName    []string `json:"KaName"`    //指定卡号, 如果指定,则生成数量无效
}

// New  制新卡
func (a *Api) New(c *gin.Context) {
	var 请求 结构请求_New
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}

	if 请求.Number <= 0 {
		response.FailWithMessage("生成数量必须大于0", c)
		return
	}
	if 请求.Number > 500 {
		response.FailWithMessage("生成数量每批最大500", c)
		return
	}
	if !Ser_Agent.Id卡类权限检测(c.GetInt("Uid"), 请求.Id) {
		response.FailWithMessage("无该卡制卡权限", c)
		return
	}

	if !Ser_KaClass.KaClassId是否存在(请求.Id) {
		response.FailWithMessage("卡类id不存在", c)
		return
	}

	数组_卡 := make([]DB.DB_Ka, 请求.Number) //make初始化,有3个元素的切片, len和cap都为3
	局_接口, ok := c.Get("局_在线信息")
	if !ok {
		response.FailWithMessage("读取缓存在线信息失败", c)
		return
	}
	局_在线信息 := 局_接口.(DB.DB_LinksToken)
	err = Ser_Ka.Ka代理批量购买(数组_卡[:], 请求.Id, 局_在线信息.Uid, 请求.AdminNote, 0, c.ClientIP())

	if err != nil {
		response.FailWithMessage("制卡失败:"+err.Error(), c)
		return
	}

	数组_卡_精简 := make([]DB_Ka_精简, 请求.Number) //make初始化,有3个元素的切片, len和cap都为3
	for 索引 := range 数组_卡_精简 {
		数组_卡_精简[索引].Name = 数组_卡[索引].Name
		数组_卡_精简[索引].Id = 数组_卡[索引].Id
		数组_卡_精简[索引].RMb = 数组_卡[索引].RMb
		数组_卡_精简[索引].VipTime = 数组_卡[索引].VipTime
		数组_卡_精简[索引].VipNumber = 数组_卡[索引].VipNumber
	}

	response.OkWithDetailed(数组_卡_精简, "制卡成功", c)
	return
}

type DB_Ka_精简 struct {
	Id        int     `json:"Id" gorm:"column:Id;primarykey"`
	Name      string  `json:"Name" gorm:"column:Name;comment:卡号"`
	VipTime   int64   `json:"VipTime" gorm:"column:VipTime;comment:增减时间秒数或点数"`
	RMb       float64 `json:"RMb" gorm:"column:RMb;type:decimal(10,2);default:0;comment:余额增减"`
	VipNumber float64 `json:"VipNumber" gorm:"column:VipNumber;type:decimal(10,2);default:0;comment:积分增减"`
}

type 结构请求_库存制卡 struct {
	Id        int    `json:"Id"`        //库存id
	Number    int    `json:"Number"`    //生成数量
	AgentNote string `json:"AgentNote"` //管理员备注
}

// New  制新卡
func (a *Api) K库存制卡(c *gin.Context) {
	var 请求 结构请求_库存制卡
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}

	数组_卡 := make([]DB.DB_Ka, 请求.Number) //make初始化,有3个元素的切片, len和cap都为3
	err = Ser_Ka.Ka代理批量库存购买(数组_卡[:], 请求.Id, 请求.Number, c.GetInt("Uid"), 请求.AgentNote, c.ClientIP())

	if err != nil {
		response.FailWithMessage("制卡失败:"+err.Error(), c)
		return
	}
	数组_卡_精简 := make([]DB_Ka_精简, 请求.Number) //make初始化,有3个元素的切片, len和cap都为3
	for 索引 := range 数组_卡_精简 {
		数组_卡_精简[索引].Name = 数组_卡[索引].Name
		数组_卡_精简[索引].Id = 数组_卡[索引].Id
		数组_卡_精简[索引].RMb = 数组_卡[索引].RMb
		数组_卡_精简[索引].VipTime = 数组_卡[索引].VipTime
		数组_卡_精简[索引].VipNumber = 数组_卡[索引].VipNumber
	}
	response.OkWithDetailed(数组_卡_精简, "制卡成功", c)
	return
}

// 批量修改状态
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
	局_权限 := false
	局_权限文本 := "卡号状态参数错误"
	switch 请求.Status {
	case 1:
		局_权限 = Ser_Agent.Id功能权限检测(c.GetInt("Uid"), DB.D代理功能_卡号解冻)
		局_权限文本 = "无卡号解冻权限,请联系上级代理授权"
	case 2:
		局_权限 = Ser_Agent.Id功能权限检测(c.GetInt("Uid"), DB.D代理功能_卡号冻结)
		局_权限文本 = "无卡号冻结权限,请联系上级代理授权"
	}
	if !局_权限 {
		response.FailWithMessage(局_权限文本, c)
		return
	}

	err = Ser_Ka.Ka修改状态(请求.Id, 请求.Status)

	if err != nil {
		response.FailWithMessage("修改失败", c)
		global.GVA_LOG.Error("修改失败:" + err.Error())
		return
	}

	response.OkWithMessage("修改成功", c)
	return
}

// G更换卡号
func (a *Api) G更换卡号(c *gin.Context) {
	var 请求 结构请求_单id
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if !Ser_Agent.Id功能权限检测(c.GetInt("Uid"), DB.D代理功能_更换卡号) {
		response.FailWithMessage("无卡号更换功能权限,请联系上级授权", c)
		return
	}

	err = Ser_Ka.Ka更换卡号(请求.Id, c.GetInt("Uid"), c.ClientIP())

	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	局_卡号详情, _ := Ser_Ka.Id取详情(请求.Id)
	response.OkWithDetailed(局_卡号详情, "更换成功", c)
	return
}

type 结构请求_批量修改状态 struct {
	Id     []int `json:"Id"`     //用户id数组
	Status int   `json:"Status"` //1 解冻 2冻结
}

// 批量修改管理员备注
func (a *Api) Set修改代理备注(c *gin.Context) {
	var 请求 结构请求_批量修改备注
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
	局_接口, ok := c.Get("局_在线信息")
	if !ok {
		response.FailWithMessage("读取缓存在线信息失败", c)
		return
	}
	局_在线信息 := 局_接口.(DB.DB_LinksToken)
	err = Ser_Ka.Ka修改代理备注(局_在线信息.User, 请求.Id, 请求.Note)

	if err != nil {
		response.FailWithMessage("修改失败", c)
		global.GVA_LOG.Error("修改失败:" + err.Error())
		return
	}

	response.OkWithMessage("修改成功", c)
	return
}

type 结构请求_批量修改备注 struct {
	Id   []int  `json:"Id"`   //用户id数组
	Note string `json:"Note"` //
}

// GetAppIdNameList 取appid和名字数组
func (a *Api) GetAppIdNameList(c *gin.Context) {

	AppIdName := Ser_AppInfo.App取map列表String()

	var Name []键值对
	局_可操作应用Id := Ser_Agent.Id取代理可操作应用AppId列表(c.GetInt("Uid"))
	for 索引 := range 局_可操作应用Id {
		Name = append(Name, 键值对{AppId: 局_可操作应用Id[索引], AppName: AppIdName[strconv.Itoa(局_可操作应用Id[索引])]})
	}

	response.OkWithDetailed(响应_AppIdNameList{AppIdName, Name}, "获取成功", c)
	return
}

type 响应_AppIdNameList struct {
	Map   map[string]string `json:"Map"`
	Array []键值对             `json:"Array"`
}

type 键值对 struct {
	AppId   int    `json:"Appid"`
	AppName string `json:"AppName"`
}
type 结构请求_充值 struct {
	Ka   string `json:"Ka"`   //卡号
	User string `json:"User"` //充值用户
}

func (a *Api) K卡号充值(c *gin.Context) {
	var 请求 结构请求_充值
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	err, _ = Ser_Ka.K卡号充值_事务(0, 请求.Ka, 请求.User, "", c.ClientIP())
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithMessage("充值成功", c)
	return
}
func (s *Api) Get卡号列表统计制卡(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get卡号列表统计制卡_代理(c), "获取成功", c)
}