package controller

import (
	"github.com/gin-gonic/gin"
	"server/new/app/controller/Common"
	"server/new/app/logic/common/userClass"
	"server/structs/Http/response"
	"sort"
	"strconv"
)

type UserClass struct {
	Common.Common
}

func NewUserClassController() *UserClass {
	return &UserClass{}
}

// GetAppIdNameList 取id和名字数组
func (C *UserClass) GetIdNameList(c *gin.Context) {
	var 请求 struct {
		Id    int `json:"Id"`
		AppId int `json:"AppId" binding:"required,min=10000"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	IdName := userClass.L_userClass.UserClass取map列表String(请求.AppId)
	var 临时Int int
	var Name []键值对

	for Key := range IdName {
		临时Int, _ = strconv.Atoi(Key)
		Name = append(Name, 键值对{Id: 临时Int, Name: IdName[Key]})
	}
	// 对 Name 数组 按键值对.Id 进行升序排序
	sort.Slice(Name, func(i, j int) bool {
		return Name[i].Id < Name[j].Id
	})

	response.OkWithDetailed(响应_AppIdNameList{IdName, Name}, "获取成功", c)
	return
}

type 键值对 struct {
	Id   int    `json:"id"`
	Name string `json:"Name"`
}

type 响应_AppIdNameList struct {
	Map   map[string]string `json:"Map"`
	Array []键值对             `json:"Array"`
}
