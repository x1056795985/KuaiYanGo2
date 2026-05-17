package controller

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
	"server/new/app/controller/Common"
	"server/new/app/logic/common/ka"
	"server/structs/Http/response"
	DB "server/structs/db"
	"strconv"
)

type KaFull struct {
	Common.Common
}

func NewKaFullController() *KaFull {
	return &KaFull{}
}

type DB_Ka_精简 struct {
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

// Info 获取卡号详情
func (C *KaFull) Info(c *gin.Context) {
	var 请求 struct {
		Id int `json:"id"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	var DB_Ka DB.DB_Ka
	err := global.GVA_DB.Model(DB.DB_Ka{}).Where("Id = ?", 请求.Id).First(&DB_Ka).Error
	if err != nil {
		response.FailWithMessage("查询详细信息失败", c)
		return
	}
	response.OkWithDetailed(DB_Ka, "获取成功", c)
}

// GetList 获取卡号列表
func (C *KaFull) GetList(c *gin.Context) {
	var 请求 struct {
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
	}
	if !C.ToJSON(c, &请求) {
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
		case 1:
			局_DB.Where("Num = NumMax")
		case 2:
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
		case 1:
			局_DB.Where("Id = ?", 请求.Keywords)
		case 2:
			局_文本数组 := utils.Z正则_取全部匹配子文本(请求.Keywords, "([A-Za-z0-9]+)")
			if len(局_文本数组) == 1 {
				局_DB.Where("Name  LIKE ?", "%"+请求.Keywords+"%")
			} else {
				局_DB.Where("Name IN ? ", 局_文本数组)
			}
		case 3:
			局_DB.Where("LOCATE(?, AdminNote)>0 ", 请求.Keywords)
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

	err := 局_DB.Count(&总数).Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&DB_Ka).Error
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		return
	}

	var AppType int = Ser_AppInfo.App取AppType(请求.AppId)
	UserClass := Ser_UserClass.UserClass取map列表Int(请求.AppId)
	KaClass := Ser_KaClass.KaClass取map列表Int(请求.AppId)

	response.OkWithDetailed(struct {
		List      interface{}    `json:"list"`
		Count     int64          `json:"count"`
		AppType   int            `json:"appType"`
		UserClass map[int]string `json:"userClass"`
		KaClass   map[int]string `json:"kaClass"`
	}{DB_Ka, 总数, AppType, UserClass, KaClass}, "获取成功", c)
}

// New 制新卡
func (C *KaFull) New(c *gin.Context) {
	var 请求 struct {
		Id        int      `json:"id"`
		Number    int      `json:"number"`
		AdminNote string   `json:"adminNote"`
		KaName    []string `json:"kaName"`
	}
	if !C.ToJSON(c, &请求) {
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

	数组_卡 := make([]DB.DB_Ka, 请求.Number)
	用户名 := Ser_LinkUser.Token取Name(c.Request.Header.Get("Token"))
	err = Ser_Ka.Ka批量创建(数组_卡[:], 请求.Id, -c.GetInt("Uid"), 用户名, 请求.AdminNote, "", 0)
	if err != nil {
		response.FailWithMessage("制卡失败:"+err.Error(), c)
		return
	}

	局_用户类型名称 := ""
	局_用户类型, ok := Ser_UserClass.Id取详情(局_卡类信息.AppId, 局_卡类信息.UserClassId)
	if ok {
		局_用户类型名称 = 局_用户类型.Name
	}

	数组_卡_精简 := make([]DB_Ka_精简, 请求.Number)
	数组_卡号 := make([]string, 请求.Number)
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
}

// BatchKaNameNew 指定卡号制卡
func (C *KaFull) BatchKaNameNew(c *gin.Context) {
	var 请求 struct {
		Id        int      `json:"id"`
		Number    int      `json:"number"`
		AdminNote string   `json:"adminNote"`
		KaName    []string `json:"kaName"`
	}
	if !C.ToJSON(c, &请求) {
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

	数组_卡 := make([]DB.DB_Ka, len(请求.KaName))
	for 索引 := range 数组_卡 {
		数组_卡[索引].Name = 请求.KaName[索引]
	}
	用户名 := Ser_LinkUser.Token取Name(c.Request.Header.Get("Token"))
	err = Ser_Ka.Ka批量创建(数组_卡[:], 局_卡类信息.Id, -c.GetInt("Uid"), 用户名, 请求.AdminNote, "", 0)
	if err != nil {
		response.FailWithMessage("导入失败:"+err.Error(), c)
		return
	}

	局_用户类型名称 := ""
	局_用户类型, ok := Ser_UserClass.Id取详情(局_卡类信息.AppId, 局_卡类信息.UserClassId)
	if ok {
		局_用户类型名称 = 局_用户类型.Name
	}

	数组_卡_精简 := make([]DB_Ka_精简, len(数组_卡))
	数组_卡号 := make([]string, len(数组_卡))
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
}

// SaveInfo 保存卡号信息
func (C *KaFull) SaveInfo(c *gin.Context) {
	var 请求 DB.DB_Ka
	if !C.ToJSON(c, &请求) {
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

	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(DB.DB_Ka{}).Where("Id= ?", 局_旧卡号信息.Id).Updates(&m).Error
		if err != nil {
			return err
		}
		if Ser_AppInfo.App是否为卡号(局_旧卡号信息.AppId) {
			err = tx.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(局_旧卡号信息.AppId)).Where("Id = ?", 局_旧卡号信息.Id).Update("Status", 请求.Status).Error
		}
		return err
	})

	if 请求.Status == 2 {
		_ = Ser_LinkUser.Set批量注销Uid数组([]int{局_旧卡号信息.Id}, 请求.AppId, Ser_LinkUser.Z注销_管理员手动注销)
	}
	if err != nil {
		response.FailWithMessage("保存失败", c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

// SetStatus 批量修改状态
func (C *KaFull) SetStatus(c *gin.Context) {
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

	err := Ser_Ka.Ka修改状态_同步卡号模式软件用户(请求.Id, 请求.Status)
	if err != nil {
		response.FailWithMessage("修改失败", c)
		return
	}
	response.OkWithMessage("修改成功", c)
}

// SetAdminNote 批量修改管理员备注
func (C *KaFull) SetAdminNote(c *gin.Context) {
	var 请求 struct {
		Id        []int  `json:"id"`
		AdminNote string `json:"adminNote"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	if len(请求.Id) == 0 {
		response.FailWithMessage("Id数组为空", c)
		return
	}
	err := Ser_Ka.Ka修改管理员备注(请求.Id, 请求.AdminNote)
	if err != nil {
		response.FailWithMessage("修改失败", c)
		return
	}
	response.OkWithMessage("修改成功", c)
}

// GetKaTemplate 获取卡号生成模板
func (C *KaFull) GetKaTemplate(c *gin.Context) {
	var 请求 struct {
		AppId int `json:"appId"`
	}
	if !C.ToJSON(c, &请求) {
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
}

// SetKaTemplate 设置卡号生成模板
func (C *KaFull) SetKaTemplate(c *gin.Context) {
	var 请求 struct {
		AppId      int    `json:"appId"`
		KaTemplate string `json:"kaTemplate"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	err := Ser_UserConfig.Z置值(1, c.GetInt("Uid"), "卡号生成格式模板"+strconv.Itoa(请求.AppId), 请求.KaTemplate)
	if err != nil {
		response.FailWithMessage("修改失败", c)
		return
	}
	response.OkWithMessage("模板已保存", c)
}

// Recover 追回卡号
func (C *KaFull) Recover(c *gin.Context) {
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
	if len(请求.Id) != 1 {
		response.FailWithMessage("Id数组暂时只支持1个成员数,后续扩展中", c)
		return
	}

	err := ka.L_ka.K卡号追回(c, 请求.Id[0], c.GetString("User"))
	if err != nil {
		response.FailWithMessage("追回失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("操作成功", c)
}

// Delete 批量删除卡号
func (C *KaFull) Delete(c *gin.Context) {
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

	var db = global.GVA_DB
	影响行数 := db.Model(DB.DB_Ka{}).Where("Id IN ? ", 请求.Id).Delete("").RowsAffected
	if db.Error != nil {
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
}

// DeleteBatch 维护删除耗尽卡号
func (C *KaFull) DeleteBatch(c *gin.Context) {
	var 请求 struct {
		AppId int `json:"appId"`
		Type  int `json:"type"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	if !Ser_AppInfo.AppId是否存在(请求.AppId) {
		response.FailWithMessage("AppId错误", c)
		return
	}
	var 局_row int64
	var err error
	switch 请求.Type {
	default:
		response.FailWithMessage("维护类型错误", c)
		return
	case 1:
		局_row, err = ka.L_ka.S删除耗尽次数卡号(c, 请求.AppId)
	}
	if err != nil {
		response.FailWithMessage("操作失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("操作成功"+strconv.Itoa(int(局_row)), c)
}
