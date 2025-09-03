package checkTaskLog

import (
	"EFunc/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"server/global"
	"server/new/app/logic/webUser/user"
	"server/new/app/models/constant"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
	"strconv"
	"time"
)

var L_checkTaskLog checkTaskLog

func init() {
	L_checkTaskLog = checkTaskLog{}
	user.L_user.Z邀请注册成功后处理(L_checkTaskLog.T邀请注册成功后处理)
}

type checkTaskLog struct {
}

func (j *checkTaskLog) T邀请注册成功后处理(c *gin.Context, AppId, 邀请人, 被邀请人 int, Referer string) {
	var err error
	var info = struct {
		AppPromotionConfigs []dbm.DB_AppPromotionConfig
		AppPromotionConfig  dbm.DB_AppPromotionConfig
		CheckInInfo         dbm.DB_CheckInInfo
		checkInUser         dbm.DB_CheckInUser
		user                DB.DB_User
	}{}

	db := *global.GVA_DB

	info.AppPromotionConfigs, err = service.NewAppPromotionConfig(c, &db).Infos(
		map[string]interface{}{
			"appId":         AppId,
			"promotionType": constant.H活动类型_签到,
		})
	if err != nil && err.Error() != "record not found" {
		return
	}

	局_当前时间戳 := time.Now().Unix()
	for i := range info.AppPromotionConfigs {
		if info.AppPromotionConfigs[i].StartTime < 局_当前时间戳 && info.AppPromotionConfigs[i].EndTime > 局_当前时间戳 {
			info.AppPromotionConfig = info.AppPromotionConfigs[i]
			break
		}
	}
	if info.AppPromotionConfig.Id == 0 {
		//没有开启活动
		return
	}
	info.CheckInInfo, err = service.NewCheckInInfo(c, &db).Info(info.AppPromotionConfig.TypeAssociatedId)
	if err != nil {
		//读取失败跳过  不增加积分也跳过
		return
	}

	局_唯一标记 := "uid" + strconv.Itoa(被邀请人)
	_, err = service.NewCheckInTaskLog(c, &db).Info(AppId, 邀请人, 2, 局_唯一标记)
	if err == nil {
		//已经领取过 分享任务奖励了
		return
	}

	info.user, err = service.NewUser(c, &db).Info(被邀请人)
	if err != nil {
		return
	}

	//插入成功,发放奖励
	//需要事务操作   增加签到积分记录, 更新用户签到分
	err = db.Transaction(func(tx *gorm.DB) error {
		//插入记录
		_, err = service.NewCheckInTaskLog(c, &db).Create(&dbm.DB_CheckInTaskLog{
			AppId:     AppId,
			UserId:    邀请人,
			Task:      2,
			Day:       局_唯一标记,
			CreatedAt: 局_当前时间戳,
			UpdatedAt: 局_当前时间戳,
		})
		if err != nil {
			//插入失败,被唯一索引限制
			return err
		}

		// 加锁重新查签到分
		err = tx.Model(dbm.DB_CheckInUser{}).Clauses(clause.Locking{Strength: "UPDATE"}).Where("appId = ?", AppId).Where("userId = ?", 邀请人).First(&info.checkInUser).Error
		if err != nil {
			return err
		}

		_, err = service.NewCheckInUser(c, tx).UpdateMap([]int{info.checkInUser.Id}, map[string]interface{}{
			"checkInScore": info.checkInUser.CheckInScore + info.CheckInInfo.ShareGivePoints,
		})
		if err != nil {
			return err
		}
		//增加签到积分记录
		_, err = service.NewCheckInScoreLog(c, tx).Create(&dbm.DB_CheckInScoreLog{
			Id:           0,
			AppId:        AppId,
			UserId:       邀请人,
			CreatedAt:    time.Now().Unix(),
			Number:       int64(info.CheckInInfo.InviteGivePoints),
			Msg:          "成功邀请好友" + utils.W文本_去除敏感信息(info.user.User),
			NumberBefore: info.checkInUser.CheckInScore,
			NumberAfter:  info.checkInUser.CheckInScore + info.CheckInInfo.ShareGivePoints,
		})

		return err
	})

	if err != nil {
		global.GVA_LOG.Error("签到邀请奖励发放触发异常", zap.Any("err", err))
		return
	}

	return
}
