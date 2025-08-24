package controller

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/controller/Common/response"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
	"time"
)

type CpsUser struct {
	Common.Common
}

func NewCpsUserController() *CpsUser {
	return &CpsUser{}
}

func (C *CpsUser) Info(c *gin.Context) {
	var err error
	var info = struct {
		appInfo  DB.DB_AppInfo
		likeInfo DB.DB_LinksToken
		cpsUser  dbm.DB_CpsUser
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	//查询是否拥有邀请人   如果已设置过,需要删除,因为有唯一索引
	tx := *global.GVA_DB
	info.cpsUser, err = service.NewCpsUser(c, &tx).Info(info.likeInfo.Uid)
	//判断是否存在,如果不存在,插入默认数据
	if err != nil && err.Error() == "record not found" {
		info.cpsUser.Id = info.likeInfo.Uid
		info.cpsUser.CreatedAt = time.Now().Unix()
		info.cpsUser.UpdatedAt = info.cpsUser.CreatedAt
		_, err = service.NewCpsUser(c, &tx).Create(&info.cpsUser)
	}

	response.OkWithData(c, info.cpsUser)
}
