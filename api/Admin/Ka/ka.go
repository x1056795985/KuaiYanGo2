package Ka

import (
	"EFunc/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_Ka"
	"server/Service/Ser_KaClass"
	"server/Service/Ser_LinkUser"
	"server/Service/Ser_Log"
	"server/Service/Ser_UserClass"
	"server/Service/Ser_UserConfig"
	"server/global"
	"server/new/app/logic/common/ka"
	"server/new/app/logic/common/log"
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

	err = global.GVA_DB.Model(DB.DB_Ka{}).Where("Id = ?", 请求.Id).First(&DB_Ka).Error
	// 没查到数据
	if err != nil {
		response.FailWithMessage("查询详细信息失败", c)
		return
	}

	response.OkWithDetailed(DB_Ka, "获取成功", c)
	return
}

type 结构请求_单id struct {
	Id int `json:"Id"`
}

// save 保存
func (a *Api) SaveKa信息(c *gin.Context) {
	var 请求 DB.DB_Ka
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	局_旧卡号信息, err2 := Ser_Ka.Id取详情(请求.Id)
	if err2 != nil {
		response.FailWithMessage("卡号不存在", c)
		return
	}

	m := map[string]interface{}{
		"Status":      请求.Status,
		"Num":         请求.Num,
		"AdminNote":   请求.AdminNote,
		"AgentNote":   请求.AgentNote,
		"VipTime":     请求.VipTime,
		"InviteCount": 请求.InviteCount,
		"RMb":         请求.RMb,
		"VipNumber":   请求.VipNumber,
		"UserClassId": 请求.UserClassId,
		"NoUserClass": 请求.NoUserClass,
	}

	//卡号模式 冻结状态关联,所以需要事务保证
	//开启事务执行
	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {

		err = tx.Model(DB.DB_Ka{}).Where("Id= ?", 局_旧卡号信息.Id).Updates(&m).Error
		if err != nil {
			return err
		}
		if Ser_AppInfo.App是否为卡号(局_旧卡号信息.AppId) {
			err = tx.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(局_旧卡号信息.AppId)).Where("Id = ?", 局_旧卡号信息.Id).Update("Status", 请求.Status).Error
		}
		return err //出错就返回并回滚
	})

	//如果是冻结同时注销在线的uid
	if 请求.Status == 2 {
		_ = Ser_LinkUser.Set批量注销Uid数组([]int{局_旧卡号信息.Id}, 请求.AppId, Ser_LinkUser.Z注销_管理员手动注销)
	}
	if err != nil {
		response.FailWithMessage("保存失败", c)
		return
	}
	response.OkWithMessage("保存成功", c)
	return
}

type 结构请求_GetKaList struct {
	AppId        int      `json:"AppId"`        // Appid 必填
	Page         int      `json:"Page"`         // 页
	Status       int      `json:"Status"`       // 状态
	RegisterTime []string `json:"RegisterTime"` // 制卡开始时间 制卡结束时间
	UseTime      []string `json:"UseTime"`      // 制卡开始时间 制卡结束时间
	KaClassId    int      `json:"KaClassId"`    // 卡类id
	Num          int      `json:"Num"`          // 卡使情况
	Size         int      `json:"Size"`         // 页数量
	Type         int      `json:"Type"`         // 关键字类型  1 id 2 用户名 3绑定信息 4 动态标签
	Keywords     string   `json:"Keywords"`     // 关键字
	Order        int      `json:"Order"`        // 0 倒序 1 正序
}

// GetKaList
func (a *Api) GetKaList(c *gin.Context) {
	var 请求 结构请求_GetKaList
	//{"AppId":2,"Type":2,"Size":10,"Page":1,"Status":1,"keywords":"1"}
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	if 请求.AppId < 10000 && 请求.AppId != 0 {
		response.FailWithMessage("AppId请输>=10000的整数", c)
		return
	}

	var DB_Ka []DB.DB_Ka
	var 总数 int64

	局_DB := global.GVA_DB.Model(DB.DB_Ka{})
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
		制卡开始时间, _ := strconv.ParseInt(请求.RegisterTime[0], 10, 64)
		制卡结束时间, _ := strconv.ParseInt(请求.RegisterTime[1], 10, 64)
		局_DB.Where("RegisterTime > ?", 制卡开始时间).Where("RegisterTime < ?", 制卡结束时间+86400)
	}
	if 请求.UseTime != nil && len(请求.UseTime) == 2 && 请求.UseTime[0] != "" && 请求.UseTime[1] != "" {
		使用开始时间, _ := strconv.ParseInt(请求.UseTime[0], 10, 64)
		使用结束时间, _ := strconv.ParseInt(请求.UseTime[1], 10, 64)
		局_DB.Where("UseTime > ?", 使用开始时间).Where("UseTime < ?", 使用结束时间+86400)
	}

	if 请求.KaClassId != 0 {
		局_DB.Where("KaClassId = ?", 请求.KaClassId)
	}

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //id
			局_DB.Where("Id = ?", 请求.Keywords)
		case 2: //卡号
			局_文本数组 := utils.Z正则_取全部匹配子文本(请求.Keywords, "([A-Za-z0-9]+)")
			if len(局_文本数组) == 1 {
				局_DB.Where("Name  LIKE ?", "%"+请求.Keywords+"%")
			} else {
				局_DB.Where("Name IN ? ", 局_文本数组)
			}
		case 3: //管理员备注
			局_DB.Where("LOCATE(?, AdminNote)>0 ", 请求.Keywords)
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

	//Count(&总数) 必须放在where 后面 不然值会被清0
	err = 局_DB.Count(&总数).Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&DB_Ka).Error
	//fmt.Println("用户总数%d", 总数, DB_LinksToken)
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		global.GVA_LOG.Error("GetKaList:" + err.Error())
		return
	}
	var AppType int
	AppType = Ser_AppInfo.App取AppType(请求.AppId)
	UserClass := Ser_UserClass.UserClass取map列表Int(请求.AppId)
	KaClass := Ser_KaClass.KaClass取map列表Int(请求.AppId)

	response.OkWithDetailed(结构响应_GetKaList{DB_Ka, 总数, AppType, UserClass, KaClass}, "获取成功", c)
	return
}

type 结构响应_GetKaList struct {
	List      interface{}    `json:"List"`      // 列表
	Count     int64          `json:"Count"`     // 总数
	AppType   int            `json:"AppType"`   //
	UserClass map[int]string `json:"UserClass"` //
	KaClass   map[int]string `json:"KaClass"`   //
}

// Del批量删除
func (a *Api) Delete(c *gin.Context) {
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
	影响行数 = db.Model(DB.DB_Ka{}).Where("Id IN ? ", 请求.Id).Delete("").RowsAffected
	if db.Error != nil {
		response.FailWithMessage("删除失败", c)
		return
	}

	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
	return
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

	err = ka.L_ka.K卡号追回(c, 请求.Id[0], c.GetString("User"))

	if err != nil {
		response.FailWithMessage("追回失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("操作成功", c)
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

	局_卡类信息, err := Ser_KaClass.KaClass取详细信息(请求.Id)
	if err != nil {
		response.FailWithMessage("卡类id不存在", c)
		return
	}
	if 请求.Number <= 0 {
		response.FailWithMessage("生成数量必须大于0", c)
		return
	}
	if 请求.Number > 2600 {
		response.FailWithMessage("生成数量每批最大2600", c)
		return
	}

	数组_卡 := make([]DB.DB_Ka, 请求.Number) //make初始化,有3个元素的切片, len和cap都为3

	用户名 := Ser_LinkUser.Token取Name(c.Request.Header.Get("Token"))
	err = Ser_Ka.Ka批量创建(数组_卡[:], 请求.Id, 用户名, 请求.AdminNote, "", 0)

	if err != nil {
		response.FailWithMessage("制卡失败:"+err.Error(), c)
		return
	}
	局_用户类型名称 := ""
	局_用户类型, ok := Ser_UserClass.Id取详情(局_卡类信息.AppId, 局_卡类信息.UserClassId)
	if ok {
		局_用户类型名称 = 局_用户类型.Name
	}
	数组_卡_精简 := make([]DB_Ka_精简, 请求.Number) //make初始化,有3个元素的切片, len和cap都为3
	数组_卡号 := make([]string, 请求.Number)     //make初始化,有3个元素的切片, len和cap都为3
	for 索引 := range 数组_卡_精简 {
		数组_卡号[索引] = 数组_卡[索引].Name
		数组_卡_精简[索引].Name = 数组_卡[索引].Name
		数组_卡_精简[索引].Id = 数组_卡[索引].Id
		数组_卡_精简[索引].RMb = 数组_卡[索引].RMb
		数组_卡_精简[索引].VipTime = 数组_卡[索引].VipTime
		数组_卡_精简[索引].VipNumber = 数组_卡[索引].VipNumber
		数组_卡_精简[索引].UserClassId = 数组_卡[索引].UserClassId
		数组_卡_精简[索引].UserClassName = 局_用户类型名称
		数组_卡_精简[索引].Num = 数组_卡[索引].Num
		数组_卡_精简[索引].MaxOnline = 数组_卡[索引].MaxOnline
		数组_卡_精简[索引].RegisterTime = 数组_卡[索引].RegisterTime
	}

	response.OkWithDetailed(数组_卡_精简, "制卡成功", c)

	局_文本 := fmt.Sprintf("新制卡号应用:%s,卡类:%s,批次id:{{批次id}}({{卡号索引}}/%d)", Ser_AppInfo.App取AppName(数组_卡[0].AppId), Ser_KaClass.Id取Name(数组_卡[0].KaClassId), 请求.Number)
	go Ser_Log.Log_写卡号操作日志(用户名, c.ClientIP(), 局_文本, 数组_卡号, 1, 4)

	return
}

// New  批量指定卡号制新卡
func (a *Api) BatchKaNameNew(c *gin.Context) {
	var 请求 结构请求_New
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}

	局_卡类信息, err := Ser_KaClass.KaClass取详细信息(请求.Id)

	if err != nil {
		response.FailWithMessage("卡类id不存在", c)
		return
	}
	if len(请求.KaName) <= 0 {
		response.FailWithMessage("导入卡号数组数量必须大于0", c)
		return
	}
	if len(请求.KaName) > 1000 {
		response.FailWithMessage("导入数量每批最大1000", c)
		return
	}

	数组_卡 := make([]DB.DB_Ka, len(请求.KaName)) //make初始化,有3个元素的切片, len和cap都为3
	for 索引, _ := range 数组_卡 {
		数组_卡[索引].Name = 请求.KaName[索引]
	}

	用户名 := Ser_LinkUser.Token取Name(c.Request.Header.Get("Token"))
	err = Ser_Ka.Ka批量创建(数组_卡[:], 局_卡类信息.Id, 用户名, 请求.AdminNote, "", 0)

	if err != nil {
		response.FailWithMessage("导入失败:"+err.Error(), c)
		return
	}
	局_用户类型名称 := ""
	局_用户类型, ok := Ser_UserClass.Id取详情(局_卡类信息.AppId, 局_卡类信息.UserClassId)
	if ok {
		局_用户类型名称 = 局_用户类型.Name
	}

	数组_卡_精简 := make([]DB_Ka_精简, len(数组_卡)) //make初始化,有3个元素的切片, len和cap都为3
	数组_卡号 := make([]string, len(数组_卡))     //make初始化,有3个元素的切片, len和cap都为3

	for 索引 := range 数组_卡_精简 {
		数组_卡号[索引] = 数组_卡[索引].Name
		数组_卡_精简[索引].Name = 数组_卡[索引].Name
		数组_卡_精简[索引].Id = 数组_卡[索引].Id
		数组_卡_精简[索引].RMb = 数组_卡[索引].RMb
		数组_卡_精简[索引].VipTime = 数组_卡[索引].VipTime
		数组_卡_精简[索引].VipNumber = 数组_卡[索引].VipNumber
		数组_卡_精简[索引].UserClassId = 数组_卡[索引].UserClassId
		数组_卡_精简[索引].UserClassName = 局_用户类型名称
		数组_卡_精简[索引].Num = 数组_卡[索引].Num
		数组_卡_精简[索引].MaxOnline = 数组_卡[索引].MaxOnline
		数组_卡_精简[索引].RegisterTime = 数组_卡[索引].RegisterTime

	}

	response.OkWithDetailed(数组_卡_精简, "导入成功", c)

	局_文本 := fmt.Sprintf("导入卡号应用:%s,卡类:%s,批次id:{{批次id}}({{卡号索引}}/%d)", Ser_AppInfo.App取AppName(数组_卡[0].AppId), Ser_KaClass.Id取Name(数组_卡[0].KaClassId), len(数组_卡号))
	go Ser_Log.Log_写卡号操作日志(用户名, c.ClientIP(), 局_文本, 数组_卡号, 1, 4)

	return
}

type DB_Ka_精简 struct {
	Id            int     `json:"Id" gorm:"column:Id;primarykey"`
	Name          string  `json:"Name" gorm:"column:Name;comment:卡号"`
	VipTime       int64   `json:"VipTime" gorm:"column:VipTime;comment:增减时间秒数或点数"`
	RMb           float64 `json:"RMb" gorm:"column:RMb;type:decimal(10,2);default:0;comment:余额增减"`
	VipNumber     float64 `json:"VipNumber" gorm:"column:VipNumber;type:decimal(10,2);default:0;comment:积分增减"`
	UserClassId   int     `json:"UserClassId" gorm:"column:UserClassId;comment:用户分类id"`
	UserClassName string  `json:"UserClassName"`
	Num           int     `json:"Num" gorm:"column:Num;comment:可以充值次数"`
	MaxOnline     int     `json:"MaxOnline" gorm:"column:MaxOnline;comment:最大在线数"` //修改可以修改App最大在线数量
	RegisterTime  int64   `json:"RegisterTime" `                                   //制卡时间
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

	if 请求.Status != 1 && 请求.Status != 2 {
		response.FailWithMessage("修改失败:Status状态代码错误", c)
		return
	}

	err = Ser_Ka.Ka修改状态_同步卡号模式软件用户(请求.Id, 请求.Status)
	if err != nil {
		response.FailWithMessage("修改失败", c)
		log.L_log.S上报异常("修改失败:" + err.Error())
		return
	}

	response.OkWithMessage("修改成功", c)
	return
}

type 结构请求_批量修改状态 struct {
	Id     []int `json:"Id"`     //用户id数组
	Status int   `json:"Status"` //1 解冻 2冻结
}

// 批量修改管理员备注
func (a *Api) Set修改管理员备注(c *gin.Context) {
	var 请求 结构请求_批量修改管理员备注
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

	err = Ser_Ka.Ka修改管理员备注(请求.Id, 请求.AdminNote)

	if err != nil {
		response.FailWithMessage("修改失败", c)
		log.L_log.S上报异常("修改失败:" + err.Error())
		return
	}

	response.OkWithMessage("修改成功", c)
	return
}

type 结构请求_Set修改卡号生成模板 struct {
	AppId      int    `json:"AppId"`
	KaTemplate string `json:"KaTemplate"`
}

func (a *Api) Set修改卡号生成模板(c *gin.Context) {
	var 请求 结构请求_Set修改卡号生成模板
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	err = Ser_UserConfig.Z置值(1, c.GetInt("Uid"), "卡号生成格式模板"+strconv.Itoa(请求.AppId), 请求.KaTemplate)
	if err != nil {
		response.FailWithMessage("修改失败", c)
		log.L_log.S上报异常("修改失败:" + err.Error())
		return
	}
	response.OkWithMessage("模板已保存", c)
	return
}
func (a *Api) Q取卡号生成模板(c *gin.Context) {
	var 请求 结构请求_Set修改卡号生成模板
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	模板 := Ser_UserConfig.Q取值(1, c.GetInt("Uid"), "卡号生成格式模板"+strconv.Itoa(请求.AppId))
	if 模板 == "" {
		模板 = "卡号:{Name} "
		if Ser_AppInfo.App是否为计点(请求.AppId) {
			模板 += "点数"
		} else {
			模板 += "时间"
		}
		模板 += ":{VipTime} 积分:{VipTime} 软件:{AppName} 余额:{RMb} 积分:{VipNumber}"
	}
	response.OkWithData(模板, c)
	return
}

type 结构请求_批量修改管理员备注 struct {
	Id        []int  `json:"Id"`        //用户id数组
	AdminNote string `json:"AdminNote"` //
}
type 结构请求_批量维护 struct {
	AppId int `json:"AppId"` //用户id数组
	Type  int `json:"Type"`  //1删除耗尽次数卡号
}

// 批量维护
func (a *Api) Set批量维护_删除用户(c *gin.Context) {
	var 请求 结构请求_批量维护
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if !Ser_AppInfo.AppId是否存在(请求.AppId) {
		response.FailWithMessage("AppId错误", c)
		return
	}
	var 局_row int64
	switch 请求.Type {
	default:
		response.FailWithMessage("维护类型错误", c)
		return
	case 1: //删除耗尽次数
		局_row, err = ka.L_ka.S删除耗尽次数卡号(c, 请求.AppId)
	}

	if err != nil {
		response.FailWithMessage("操作失败:"+err.Error(), c)
		log.L_log.S上报异常("操作失败:" + err.Error())
		return
	}

	response.OkWithMessage("操作成功"+strconv.Itoa(int(局_row)), c)
	return
}
