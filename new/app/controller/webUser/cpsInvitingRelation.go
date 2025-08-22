package controller

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/controller/Common/response"
	"server/new/app/logic/webUser/cpsInvitingRelation"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
)

type CpsInvitingRelation struct {
	Common.Common
}

func NewCpsInvitingRelationController() *CpsInvitingRelation {
	return &CpsInvitingRelation{}
}

// 设置邀请关系
func (C *CpsInvitingRelation) Set(c *gin.Context) {
	var 请求 struct {
		PromotionCode int `json:"promotionCode"  binding:"required" zh:"邀请代码"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	var err error
	var info = struct {
		appInfo  DB.DB_AppInfo
		likeInfo DB.DB_LinksToken
		上级       dbm.DB_CpsInvitingRelation
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	//查询是否拥有邀请人   如果已设置过,需要删除,因为有唯一索引
	tx := *global.GVA_DB
	info.上级, _, err = service.NewCpsInvitingRelation(c, &tx).Q取归属邀请人(info.appInfo.AppId, info.likeInfo.Uid)
	if info.上级.Id == 0 { //没有推荐人,才设置
		if info.likeInfo.Uid == 请求.PromotionCode {
			response.FailWithMessage(c, "不能邀请自己")
			return
		}

		err = cpsInvitingRelation.L_CpsInvitingRelation.S设置邀请人(c, info.appInfo.AppId, 请求.PromotionCode, info.likeInfo.Uid, c.GetHeader("Referer"))
		if err != nil {
			response.FailWithMessage(c, err.Error())
			return
		}
	}
	response.Ok(c)
}
func (C *CpsInvitingRelation) Get(c *gin.Context) {
	var info = struct {
		appInfo  DB.DB_AppInfo
		likeInfo DB.DB_LinksToken
		Y邀请关系    dbm.DB_CpsInvitingRelation
	}{}

	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	tx := *global.GVA_DB
	info.Y邀请关系, _, _ = service.NewCpsInvitingRelation(c, &tx).Q取归属邀请人(info.appInfo.AppId, info.likeInfo.Uid)

	response.OkWithData(c, gin.H{
		"isInput":   info.appInfo.AppType <= 2, //只有账号模式才需要填写
		"inviterId": info.Y邀请关系.InviterId,
		"createdAt": info.Y邀请关系.CreatedAt,
		"updatedAt": info.Y邀请关系.UpdatedAt,
	})
}

func (C *CpsInvitingRelation) GetInvitingList(c *gin.Context) {
	var info = struct {
		appInfo  DB.DB_AppInfo
		likeInfo DB.DB_LinksToken
		Y邀请关系    []dbm.DB_CpsInvitingRelation
	}{}

	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	tx := *global.GVA_DB
	info.Y邀请关系, _ = service.NewCpsInvitingRelation(c, &tx).Q取所有被邀请人(info.appInfo.AppId, info.likeInfo.Uid, 50)

	var err error
	var ids []int
	for _, v := range info.Y邀请关系 {
		ids = append(ids, v.InviteeId)
	}
	var 邀请人 []DB.DB_User
	邀请人, err = service.NewUser(c, &tx).Infos(map[string]interface{}{"Id": ids})
	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}

	var 响应data = make([]struct {
		User         string `json:"user"`
		InvitingTime int64  `json:"invitingTime"`
	}, 0, len(info.Y邀请关系))

	键值对 := make(map[int]string) // 初始化 map
	for _, v := range 邀请人 {
		键值对[v.Id] = v.User
	}
	for _, v := range info.Y邀请关系 {
		// 安全访问示例（可选）：
		if user, exists := 键值对[v.InviteeId]; exists {
			响应data = append(响应data, struct {
				User         string `json:"user"`
				InvitingTime int64  `json:"invitingTime"`
			}{
				User:         W文本_去除敏感信息(user),
				InvitingTime: v.CreatedAt,
			})

		}
	}

	response.OkWithData(c, 响应data)
}
