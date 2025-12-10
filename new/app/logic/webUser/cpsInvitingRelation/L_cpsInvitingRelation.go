package cpsInvitingRelation

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"server/global"
	"server/new/app/logic/common/ka"
	"server/new/app/logic/common/log"
	"server/new/app/logic/webUser/cps"
	"server/new/app/logic/webUser/user"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
	"strconv"
	"time"
)

var L_CpsInvitingRelation appUser

func init() {
	L_CpsInvitingRelation = appUser{}
}

type appUser struct {
}

// 四舍五入  索引越小,代理级别越靠下  cps邀请专用
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
	info.上级, info.上上级, err = service.NewCpsInvitingRelation(c, tx).Q取归属邀请人(AppId, 邀请人)
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
	if info.上级.Id > 0 { //如果有就加上,如果没有就算了
		info.插入数据 = append(info.插入数据, dbm.DB_CpsInvitingRelation{
			CreatedAt:    局_time,
			UpdatedAt:    局_time,
			InviterId:    info.上级.InviterId,
			InviteeAppId: AppId,
			InviteeId:    被邀请人,
			Level:        2,
			Status:       1,
			Referer:      Referer,
		})
	}

	err = tx.Create(&info.插入数据).Error
	if err == nil {
		user.L_user.T邀请注册成功后处理(c, AppId, 邀请人, 被邀请人, Referer)
	}
	return
}
func (j *appUser) F发放被邀奖励卡(c *gin.Context, AppId int, Uid int) (err error) {

	var info struct {
		AppUser DB.DB_AppUser
		cpsInfo dbm.DB_CpsInfo

		LogMoney     []DB.DB_LogMoney
		LogVipNumber []DB.DB_LogVipNumber
	}
	var tx *gorm.DB
	if tempObj, ok := c.Get("tx"); ok {
		tx = tempObj.(*gorm.DB)
	} else {
		db := *global.GVA_DB
		tx = &db
	}

	info.AppUser, err = service.NewAppUser(c, tx, AppId).InfoUid(Uid)
	if err != nil {
		return errors.New("用户不存在")
	}
	info.cpsInfo = cps.L_cps.Q开启中cps活动(c, AppId)
	if info.cpsInfo.Id <= 0 {
		return errors.New("没有开启的CPS活动")
	}
	//开始充值 卡号
	if info.cpsInfo.BindGiveKaClassId > 0 {
		// 子查询获取所有软件用户的Uid 在修改卡号
		if err = ka.L_ka.K卡类直冲_事务(c, info.cpsInfo.BindGiveKaClassId, Uid); err != nil {
			return err
		}
	}

	if 临时数据, ok := c.Get("logMoney"); ok { //判断是否有rmb充值的日志
		info.LogMoney = append(info.LogMoney, 临时数据.(DB.DB_LogMoney))
		info.LogMoney[len(info.LogMoney)-1].Note = "归属代理送卡,卡类id:" + strconv.Itoa(info.cpsInfo.BindGiveKaClassId) + info.LogMoney[len(info.LogMoney)-1].Note
	}

	if 临时数据, ok := c.Get("logVipNumber"); ok { //判断是否有积分充值的日志
		info.LogVipNumber = append(info.LogVipNumber, 临时数据.(DB.DB_LogVipNumber))
		info.LogVipNumber[len(info.LogVipNumber)-1].Note = "归属代理送卡,卡类id:" + strconv.Itoa(info.cpsInfo.BindGiveKaClassId) + info.LogVipNumber[len(info.LogVipNumber)-1].Note
	}
	//最后写出日志
	if err = log.L_log.S输出日志(c, info.LogMoney); err != nil {
		global.GVA_LOG.Error("输出日志失败!", zap.Any("err", err))
	}
	if err = log.L_log.S输出日志(c, info.LogVipNumber); err != nil {
		global.GVA_LOG.Error("输出日志失败!", zap.Any("err", err))
	}

	return
}
