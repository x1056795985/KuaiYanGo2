package controller

import (
	"github.com/gin-gonic/gin"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_UserClass"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/logic/common/agent"
	dbm "server/new/app/models/db"
	"server/new/app/models/request"
	"server/new/app/service"
	"server/structs/Http/response"
	"server/utils"
	"strconv"
)

// KaClass 卡类列表管理
type KaClass struct {
	Common.Common
}

func NewKaClassController() *KaClass {
	return &KaClass{}
}

// 请求结构体（与旧架构完全一致的JSON参数名）
type 请求_KaClassGetInfo struct {
	Id int `json:"id"`
}

type 请求_KaClassGetList struct {
	AppId    int    `json:"appId"`    // Appid 必填
	Page     int    `json:"page"`     // 页
	Size     int    `json:"size"`     // 页数量
	Type     int    `json:"type"`     // 关键字类型  1 id 2 卡类名称
	Keywords string `json:"keywords"` // 关键字
	Order    int    `json:"order"`    // 0 倒序 1 正序
}

type 请求_KaClassDelete struct {
	Id    []int `json:"id"`    //id数组
	AppId int   `json:"appId"`
}

type 请求_KaClassGetListAll struct {
	AppId int `json:"appId"`
}

type 响应_KaClassGetList struct {
	List      interface{}    `json:"list"`      // 列表
	Count     int64          `json:"count"`     // 总数
	UserClass map[int]string `json:"userClass"` //
	AppType   int            `json:"appType"`   //
}

// Info 获取卡类详细信息
func (C *KaClass) Info(c *gin.Context) {
	var 请求 请求_KaClassGetInfo
	if !C.ToJSON(c, &请求) {
		return
	}

	S := service.NewKaClass(c, global.GVA_DB)
	info, err := S.Info(请求.Id)
	if err != nil {
		response.FailWithMessage("查询详细信息失败", c)
		return
	}
	response.OkWithDetailed(info, "获取成功", c)
}

// GetList 获取卡类列表
func (C *KaClass) GetList(c *gin.Context) {
	var 请求 请求_KaClassGetList
	if !C.ToJSON(c, &请求) {
		return
	}

	if 请求.AppId < 10000 {
		response.FailWithMessage("AppId请输>=10000的整数", c)
		return
	}

	S := service.NewKaClass(c, global.GVA_DB)

	listReq := request.List{
		Page:     请求.Page,
		Size:     请求.Size,
		Type:     请求.Type,
		Keywords: 请求.Keywords,
		Order:    请求.Order,
	}

	总数, dataList, err := S.GetList(listReq, 请求.AppId, nil)
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		return
	}

	AppType := Ser_AppInfo.App取AppType(请求.AppId)
	UserClass := Ser_UserClass.UserClass取map列表Int(请求.AppId)

	response.OkWithDetailed(响应_KaClassGetList{dataList, 总数, UserClass, AppType}, "获取成功", c)
}

// Delete 批量删除卡类
func (C *KaClass) Delete(c *gin.Context) {
	var 请求 请求_KaClassDelete
	if !C.ToJSON(c, &请求) {
		return
	}

	if len(请求.Id) == 0 {
		response.FailWithMessage("Id数组为空", c)
		return
	}

	S := service.NewKaClass(c, global.GVA_DB)
	影响行数, err := S.Delete(请求.Id)
	if err != nil {
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
}

// SaveInfo 保存卡类信息
func (C *KaClass) SaveInfo(c *gin.Context) {
	var 请求 dbm.DB_KaClass
	if !C.ToJSON(c, &请求) {
		return
	}

	if 请求.Id <= 0 {
		response.FailWithMessage("Id错误", c)
		return
	}
	if len(请求.Note) > 400 {
		response.FailWithMessage("备注过长,请减少备注长度", c)
		return
	}
	if 请求.KaLength-len(请求.Prefix) < 10 {
		response.FailWithMessage("制卡可随机字符长度小于10,请增加卡长度或减少前缀长度", c)
		return
	}
	if 请求.KaLength+len(请求.Prefix) > 191 {
		response.FailWithMessage("制卡可随机字符长度最大191,请减少卡长度或减少前缀长度", c)
		return
	}
	if 请求.VipNumber < 0 || 请求.VipTime < 0 || 请求.InviteCount < 0 || 请求.Num < 0 {
		response.FailWithMessage("值不能为小于0", c)
		return
	}
	if 请求.Money < -1 || 请求.AgentMoney < -1 {
		response.FailWithMessage("售价值不能为小于-1", c)
		return
	}

	S := service.NewKaClass(c, global.GVA_DB)
	if !S.IsIdExists(请求.Id) {
		response.FailWithMessage("卡类不存在", c)
		return
	}

	data := map[string]interface{}{
		"KaStringType": 请求.KaStringType,
		"Note":         请求.Note,
		"MaxOnline":    请求.MaxOnline,
		"Name":         请求.Name,
		"Prefix":       请求.Prefix,
		"VipTime":      请求.VipTime,
		"KaType":       请求.KaType,
		"Num":          请求.Num,
		"KaLength":     请求.KaLength,
		"NoUserClass":  请求.NoUserClass,
		"UserClassId":  请求.UserClassId,
		"AgentMoney":   请求.AgentMoney,
		"Money":        请求.Money,
		"VipNumber":    请求.VipNumber,
		"RMb":          请求.RMb,
		"InviteCount":  请求.InviteCount,
	}

	if Ser_AppInfo.App是否为卡号(请求.AppId) {
		data["Num"] = 1 //卡号类型卡只能用一次
	}

	_, err := S.Update(请求.Id, data)
	if err != nil {
		response.FailWithMessage("保存失败", c)
		return
	}

	if 请求.AgentMoney < 0 {
		agent.L_agent.D代理授权卡类Id删除(c, 请求.Id)
	}

	response.OkWithMessage("保存成功", c)
}

// New 新建卡类
func (C *KaClass) New(c *gin.Context) {
	var 请求 dbm.DB_KaClass
	if !C.ToJSON(c, &请求) {
		return
	}

	if 请求.Id > 0 {
		response.FailWithMessage("添加用户不能有id值", c)
		return
	}
	if 请求.AppId < 10000 || !Ser_AppInfo.AppId是否存在(请求.AppId) {
		response.FailWithMessage("AppId错误", c)
		return
	}
	if 请求.Name == "" {
		response.FailWithMessage("卡类名称不能为空", c)
		return
	}

	var msg string
	if 请求.Prefix != "" && !utils.Z正则_是否英数(请求.Prefix, &msg) {
		response.FailWithMessage("卡类前缀"+msg, c)
		return
	}

	if !Ser_AppInfo.AppId是否存在(请求.AppId) {
		response.FailWithMessage("AppId不存在,请先去[ 应用管理 => 应用列表 ],添加该应用信息", c)
		return
	}

	if 请求.KaLength-len(请求.Prefix) < 10 {
		response.FailWithMessage("制卡可随机字符长度小于10,请增加卡长度或减少前缀长度", c)
		return
	}

	if 请求.VipNumber < 0 || 请求.VipTime < 0 || 请求.InviteCount < 0 || 请求.Num < 0 {
		response.FailWithMessage("值不能为小于0", c)
		return
	}

	if 请求.Money < -1 || 请求.AgentMoney < -1 {
		response.FailWithMessage("售价值不能为小于-1", c)
		return
	}

	if Ser_AppInfo.App是否为卡号(请求.AppId) {
		请求.Num = 1 //卡号类型卡只能用一次
	}

	S := service.NewKaClass(c, global.GVA_DB)
	_, err := S.Create(&请求)
	if err != nil {
		response.FailWithMessage("添加失败", c)
		return
	}
	response.OkWithMessage("添加成功", c)
}

// GetListAll 按AppId获取全部卡类（不分页）
func (C *KaClass) GetListAll(c *gin.Context) {
	var 请求 请求_KaClassGetListAll
	if !C.ToJSON(c, &请求) {
		return
	}

	if 请求.AppId < 10000 {
		response.FailWithMessage("AppId请输>=10000的整数", c)
		return
	}

	S := service.NewKaClass(c, global.GVA_DB)
	dataList, err := S.GetListAll(请求.AppId)
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		return
	}
	response.OkWithDetailed(dataList, "ok", c)
}
