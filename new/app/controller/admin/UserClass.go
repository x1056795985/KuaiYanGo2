package controller

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/models/request"
	"server/new/app/service"
	"server/structs/Http/response"
	DB "server/structs/db"
	"sort"
	"strconv"
)

// UserClass 用户类型管理
type UserClass struct {
	Common.Common
}

func NewUserClassController() *UserClass {
	return &UserClass{}
}

// 请求结构体（与旧架构完全一致的JSON参数名）
type 请求_UserClassGetInfo struct {
	Id    int `json:"id"`
	AppId int `json:"appId"`
}

type 请求_UserClassGetList struct {
	AppId    int    `json:"appId"`    // Appid 必填
	Page     int    `json:"page"`     // 页
	Size     int    `json:"size"`     // 页数量
	Type     int    `json:"type"`     // 关键字类型  1 id
	Keywords string `json:"keywords"` // 关键字
	Order    int    `json:"order"`    // 0 倒序 1 正序
}

type 请求_UserClassDelete struct {
	Id    []int `json:"id"`    //用户id数组
	AppId int   `json:"appId"` // Appid 必填
}

type 响应_UserClassGetIdNameList struct {
	Map   map[string]string `json:"map"`
	Array []键值对_UserClass  `json:"array"`
}

type 键值对_UserClass struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// Info 获取用户类型详细信息
func (C *UserClass) Info(c *gin.Context) {
	var 请求 请求_UserClassGetInfo
	if !C.ToJSON(c, &请求) {
		return
	}

	S := service.NewUserClass(c, global.GVA_DB)
	info, err := S.Info(请求.Id)
	if err != nil {
		response.FailWithMessage("查询用户类型信息失败.id可能不存在", c)
		return
	}
	response.OkWithDetailed(info, "获取成功", c)
}

// GetList 获取用户类型列表
func (C *UserClass) GetList(c *gin.Context) {
	var 请求 请求_UserClassGetList
	if !C.ToJSON(c, &请求) {
		return
	}

	if 请求.AppId < 10000 {
		response.FailWithMessage("AppId请输>=10000的整数", c)
		return
	}

	S := service.NewUserClass(c, global.GVA_DB)

	listReq := request.List{
		Page:     请求.Page,
		Size:     请求.Size,
		Type:     请求.Type,
		Keywords: 请求.Keywords,
		Order:    请求.Order,
	}

	总数, dataList, err := S.GetListByAppId(请求.AppId, listReq)
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		return
	}
	response.OkWithDetailed(struct {
		List  interface{} `json:"list"`
		Count int64       `json:"count"`
	}{dataList, 总数}, "获取成功", c)
}

// Delete 批量删除用户类型
func (C *UserClass) Delete(c *gin.Context) {
	var 请求 请求_UserClassDelete
	if !C.ToJSON(c, &请求) {
		return
	}

	if len(请求.Id) == 0 {
		response.FailWithMessage("Id数组为空", c)
		return
	}
	if 请求.AppId < 10000 {
		response.FailWithMessage("AppId请输>=10000的整数", c)
		return
	}

	S := service.NewUserClass(c, global.GVA_DB)

	影响行数, err := S.DeleteByAppIdAndIds(请求.AppId, 请求.Id)
	if err != nil {
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
}

// SaveInfo 保存用户类型信息
func (C *UserClass) SaveInfo(c *gin.Context) {
	var 请求 DB.DB_UserClass
	if !C.ToJSON(c, &请求) {
		return
	}

	if 请求.Name == "" {
		response.FailWithMessage("用户名称不能为空", c)
		return
	}
	if 请求.Id <= 0 {
		response.FailWithMessage("Id错误", c)
		return
	}
	if 请求.Weight <= 0 {
		response.FailWithMessage("权重最小为1", c)
		return
	}

	S := service.NewUserClass(c, global.GVA_DB)

	if !S.IsIdExists(请求.Id) {
		response.FailWithMessage("用户类型不存在", c)
		return
	}

	if S.IsMarkExistsCount(请求.AppId, 请求.Mark, []int{请求.Id}) >= 1 {
		response.FailWithMessage("整数代号已存在", c)
		return
	}

	data := map[string]interface{}{
		"Name":   请求.Name,
		"Mark":   请求.Mark,
		"Weight": 请求.Weight,
	}
	_, err := S.Update(请求.Id, data)
	if err != nil {
		response.FailWithMessage("保存失败", c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

// New 新建用户类型
func (C *UserClass) New(c *gin.Context) {
	var 请求 DB.DB_UserClass
	if !C.ToJSON(c, &请求) {
		return
	}

	if 请求.AppId < 10000 {
		response.FailWithMessage("AppId错误", c)
		return
	}
	if 请求.Id > 0 {
		response.FailWithMessage("添加用户不能有id值", c)
		return
	}
	if 请求.Weight <= 0 {
		response.FailWithMessage("权重最小为1", c)
		return
	}

	S := service.NewUserClass(c, global.GVA_DB)

	if S.IsNameExists(请求.AppId, 请求.Name) {
		response.FailWithMessage("用户类型名称已存在", c)
		return
	}
	if S.IsMarkExists(请求.AppId, 请求.Mark) {
		response.FailWithMessage("整数代号已存在", c)
		return
	}

	_, err := S.Create(请求)
	if err != nil {
		response.FailWithMessage("添加失败", c)
		return
	}
	response.OkWithMessage("添加成功", c)
}

// GetIdNameList 取id和名字数组
func (C *UserClass) GetIdNameList(c *gin.Context) {
	var 请求 请求_UserClassGetInfo
	if !C.ToJSON(c, &请求) {
		return
	}
	if 请求.AppId < 10000 {
		response.FailWithMessage("AppId错误", c)
		return
	}

	S := service.NewUserClass(c, global.GVA_DB)

	IdName, err := S.GetIdNameList(请求.AppId)
	if err != nil {
		response.FailWithMessage("获取失败", c)
		return
	}

	var Name []键值对_UserClass
	for Key := range IdName {
		临时Int, _ := strconv.Atoi(Key)
		Name = append(Name, 键值对_UserClass{Id: 临时Int, Name: IdName[Key]})
	}
	sort.Slice(Name, func(i, j int) bool {
		return Name[i].Id < Name[j].Id
	})
	response.OkWithDetailed(响应_UserClassGetIdNameList{IdName, Name}, "获取成功", c)
}
