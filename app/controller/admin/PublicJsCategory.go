package controller

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"server/app/controller/Common"
	"server/app/global"
	"server/app/models/dbm"
	"server/app/models/old/response"
	"server/app/service"
)

type PublicJsCategoryCtrl struct {
	Common.Common
}

func NewPublicJsCategoryController() *PublicJsCategoryCtrl {
	return &PublicJsCategoryCtrl{}
}

type 请求_分类新建 struct {
	ParentId int    `json:"ParentId"`
	Name     string `json:"Name"`
	Sort     int64  `json:"Sort"`
	Note     string `json:"Note"`
}

type 请求_分类修改 struct {
	Id       int    `json:"Id"`
	ParentId int    `json:"ParentId"`
	Name     string `json:"Name"`
	Sort     int64  `json:"Sort"`
	Note     string `json:"Note"`
}

type 响应_分类详情 struct {
	dbm.DB_PublicJsCategory
	Count int64 `json:"Count"` //分类下函数数量
}

// GetList 获取分类列表(带每个分类下的函数数量统计)
func (C *PublicJsCategoryCtrl) GetList(c *gin.Context) {
	db := *global.GVA_DB
	局_service := service.NewPublicJsCategory(c, &db)
	局_分类列表 := 局_service.Q取列表()
	局_数量统计 := 局_service.P取分类函数数量()

	局_响应列表 := make([]响应_分类详情, 0, len(局_分类列表))
	for _, v := range 局_分类列表 {
		局_响应列表 = append(局_响应列表, 响应_分类详情{DB_PublicJsCategory: v, Count: 局_数量统计[v.Id]})
	}

	//CategoryId=0 即未分类的数量
	response.OkWithDetailed(gin.H{"list": 局_响应列表, "未分类Count": 局_数量统计[0]}, "获取成功", c)
}

// New 新建分类
func (C *PublicJsCategoryCtrl) New(c *gin.Context) {
	var 请求 请求_分类新建
	if !C.ToJSON(c, &请求) {
		return
	}
	请求.Name = strings.TrimSpace(请求.Name)
	if 请求.Name == "" {
		response.FailWithMessage("分类名不能为空", c)
		return
	}

	db := *global.GVA_DB
	局_service := service.NewPublicJsCategory(c, &db)
	if 局_service.Name取Id(请求.Name) != 0 {
		response.FailWithMessage("分类名已存在", c)
		return
	}
	if 请求.ParentId < 0 {
		response.FailWithMessage("上级分类Id错误", c)
		return
	}
	if 请求.ParentId > 0 && !局_service.Id是否存在(请求.ParentId) {
		response.FailWithMessage("上级分类不存在", c)
		return
	}

	err := db.Model(dbm.DB_PublicJsCategory{}).Create(&dbm.DB_PublicJsCategory{
		ParentId: 请求.ParentId,
		Name:     请求.Name,
		Sort:     请求.Sort,
		Note:     请求.Note,
	}).Error
	if err != nil {
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("添加成功", c)
}

// SaveInfo 修改分类信息
func (C *PublicJsCategoryCtrl) SaveInfo(c *gin.Context) {
	var 请求 请求_分类修改
	if !C.ToJSON(c, &请求) {
		return
	}
	请求.Name = strings.TrimSpace(请求.Name)
	if 请求.Name == "" {
		response.FailWithMessage("分类名不能为空", c)
		return
	}
	if 请求.Id <= 0 {
		response.FailWithMessage("分类Id错误", c)
		return
	}

	db := *global.GVA_DB
	局_service := service.NewPublicJsCategory(c, &db)
	if !局_service.Id是否存在(请求.Id) {
		response.FailWithMessage("分类不存在或已被删除", c)
		return
	}
	if 局_同Id := 局_service.Name取Id(请求.Name); 局_同Id != 0 && 局_同Id != 请求.Id {
		response.FailWithMessage("分类名已存在", c)
		return
	}
	if 请求.ParentId < 0 {
		response.FailWithMessage("上级分类Id错误", c)
		return
	}
	if 请求.ParentId > 0 {
		if !局_service.Id是否存在(请求.ParentId) {
			response.FailWithMessage("上级分类不存在", c)
			return
		}
		//不能把自己移动到自己或自己的子孙分类下(防止环形引用)
		if 局_service.Q是否自身或子孙分类(请求.Id, 请求.ParentId) {
			response.FailWithMessage("上级分类不能是自己或自己的子分类", c)
			return
		}
	}

	err := db.Model(dbm.DB_PublicJsCategory{}).Where("Id=?", 请求.Id).Updates(map[string]interface{}{
		"ParentId": 请求.ParentId,
		"Name":     请求.Name,
		"Sort":     请求.Sort,
		"Note":     请求.Note,
	}).Error
	if err != nil {
		response.FailWithMessage("保存失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

// Delete 删除分类(其下函数自动移入未分类)
func (C *PublicJsCategoryCtrl) Delete(c *gin.Context) {
	var 请求 struct {
		Id int `json:"Id"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	if 请求.Id <= 0 {
		response.FailWithMessage("分类Id错误", c)
		return
	}

	db := *global.GVA_DB
	局_service := service.NewPublicJsCategory(c, &db)
	if !局_service.Id是否存在(请求.Id) {
		response.FailWithMessage("分类不存在或已被删除", c)
		return
	}

	//删除分类及其整个子树,树下所有函数移入未分类
	局_移动数量, err := 局_service.D删除(请求.Id)
	if err != nil {
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功,该分类下" + strconv.FormatInt(局_移动数量, 10) + "个函数已移至未分类", c)
}

type 请求_批量移动分类 struct {
	Id         []int `json:"Id"`
	CategoryId int   `json:"CategoryId"` //0=移入未分类
}

// SetFunctionCategory 批量移动函数到指定分类
func (C *PublicJsCategoryCtrl) SetFunctionCategory(c *gin.Context) {
	var 请求 请求_批量移动分类
	if !C.ToJSON(c, &请求) {
		return
	}
	if len(请求.Id) == 0 {
		response.FailWithMessage("请先勾选要移动的函数", c)
		return
	}
	if 请求.CategoryId < 0 {
		response.FailWithMessage("分类Id错误", c)
		return
	}

	db := *global.GVA_DB
	if 请求.CategoryId > 0 && !service.NewPublicJsCategory(c, &db).Id是否存在(请求.CategoryId) {
		response.FailWithMessage("目标分类不存在", c)
		return
	}

	err := service.NewPublicJsCategory(c, &db).P批量修改分类(请求.Id, 请求.CategoryId)
	if err != nil {
		response.FailWithMessage("移动失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("移动成功,数量" + strconv.Itoa(len(请求.Id)), c)
}
