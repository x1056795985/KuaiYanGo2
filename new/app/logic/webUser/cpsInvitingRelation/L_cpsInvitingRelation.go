package cpsInvitingRelation

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/global"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
	"time"
)

var L_CpsInvitingRelation appUser

func init() {
	L_CpsInvitingRelation = appUser{}

}

type appUser struct {
}

// 四舍五入  索引越小,代理级别越靠下  代理专用
func (j *appUser) S设置邀请人(c *gin.Context, AppId, 邀请人, 被邀请人 int, Referer string) (err error) {
	var info struct {
		AppInfo DB.DB_AppInfo
		上级      dbm.DB_CpsInvitingRelation
		上上级     dbm.DB_CpsInvitingRelation
		插入数据    []dbm.DB_CpsInvitingRelation
		邀请人信息   DB.DB_User
	}
	var tx *gorm.DB
	if tempObj, ok := c.Get("tx"); ok {
		tx = tempObj.(*gorm.DB)
	} else {
		db := *global.GVA_DB
		tx = &db
	}

	if 邀请人 == 被邀请人 {
		err = errors.New("邀请人不能是自己")
		return
	}
	info.邀请人信息, err = service.NewUser(c, tx).Info(邀请人)
	if err != nil {
		err = errors.New("邀请人不存在")
		return
	}

	//查询是否拥有邀请人   如果已设置过,需要删除,因为有唯一索引
	info.上级, info.上上级, err = service.NewCpsInvitingRelation(c, tx).Q取归属邀请人(AppId, 被邀请人)
	if info.上级.Id > 0 {
		// 删除上级关系
		_ = tx.Delete(&info.上级)
	}
	if info.上上级.Id > 0 {
		// 删除上上级关系
		_ = tx.Delete(&info.上级)
	}
	_, info.上上级, err = service.NewCpsInvitingRelation(c, tx).Q取归属邀请人(AppId, 邀请人)
	局_time := time.Now().Unix()
	info.插入数据 = make([]dbm.DB_CpsInvitingRelation, 0, 2)
	info.插入数据 = append(info.插入数据, dbm.DB_CpsInvitingRelation{
		CreatedAt:    局_time,
		UpdatedAt:    局_time,
		InviterId:    邀请人,
		InviteeAppId: AppId,
		InviteeId:    被邀请人,
		Level:        1,
		Status:       1,
		Referer:      Referer,
	})
	if info.上上级.Id > 0 { //如果有就加上,如果没有就算了
		info.插入数据 = append(info.插入数据, dbm.DB_CpsInvitingRelation{
			CreatedAt:    局_time,
			UpdatedAt:    局_time,
			InviterId:    info.上上级.InviteeId,
			InviteeAppId: AppId,
			InviteeId:    被邀请人,
			Level:        2,
			Status:       1,
			Referer:      Referer,
		})
	}

	err = tx.Create(&info.插入数据).Error
	return
}
