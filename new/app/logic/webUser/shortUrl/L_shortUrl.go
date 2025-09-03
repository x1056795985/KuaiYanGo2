package shortUr

import (
	"EFunc/utils"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"server/global"
	"server/new/app/models/constant"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	"time"
)

var L_shortUr shortUr

func init() {
	L_shortUr = shortUr{}

}

type shortUr struct {
}

func (j *shortUr) D短链访问后处理(c *gin.Context, 短链信息 dbm.DB_ShortUrl) (有效数量 int) {
	tx := *global.GVA_DB
	// 跳转成功,写入数据库计数+1
	_ = service.NewShortUrl(c, &tx).ClickCountUP(短链信息.Id, 1)
	switch 短链信息.ShortType {
	case constant.H活动类型_cps:
		return j.cps分享被访问(c, 短链信息)
	case constant.H活动类型_签到:
		return j.签到分享被访问(c, 短链信息)

	}

	return
}
func (j *shortUr) cps分享被访问(c *gin.Context, 短链信息 dbm.DB_ShortUrl) (有效数量 int) {

	return
}

func (j *shortUr) 签到分享被访问(c *gin.Context, 短链信息 dbm.DB_ShortUrl) (有效数量 int) {
	var err error
	var info = struct {
		AppPromotionConfigs []dbm.DB_AppPromotionConfig
		AppPromotionConfig  dbm.DB_AppPromotionConfig
		CheckInInfo         dbm.DB_CheckInInfo
		checkInUser         dbm.DB_CheckInUser
	}{}

	db := *global.GVA_DB

	info.AppPromotionConfigs, err = service.NewAppPromotionConfig(c, &db).Infos(
		map[string]interface{}{
			"appId":         短链信息.AppId,
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
	//风控拦截,UserAgent 必须包含关键词 QQ  微信  WeChat
	局_ua := c.Request.UserAgent()
	if !utils.W文本_是否包含关键字(局_ua, "QQ") && !utils.W文本_是否包含关键字(局_ua, "WeChat") {
		//必须通过QQ或微信内置浏览器打开才可以
		//未通过风控
		return
	}
	//和创建人使用相同ip也不可以
	局_访问人ip := c.ClientIP()
	var other struct {
		Ip string `json:"ip"`
	}
	err = json.Unmarshal([]byte(短链信息.Other), &other)
	if err == nil {
		if other.Ip == 局_访问人ip {
			//不能使用相同ip
			return
		}
	}

	局_今天唯一标记 := time.Now().Format("20060102")
	_, err = service.NewCheckInTaskLog(c, &db).Info(短链信息.AppId, 短链信息.Uid, 1, 局_今天唯一标记)
	if err == nil {
		//已经领取过 分享任务奖励了
		return
	}

	//插入成功,发放奖励
	//需要事务操作   增加签到积分记录, 更新用户签到分
	err = db.Transaction(func(tx *gorm.DB) error {
		//插入记录
		_, err = service.NewCheckInTaskLog(c, &db).Create(&dbm.DB_CheckInTaskLog{
			AppId:     短链信息.AppId,
			UserId:    短链信息.Uid,
			Task:      1,
			Day:       局_今天唯一标记,
			CreatedAt: 局_当前时间戳,
			UpdatedAt: 局_当前时间戳,
		})
		if err != nil {
			//插入失败,被唯一索引限制
			return err
		}

		// 加锁重新查签到分
		err = tx.Model(dbm.DB_CheckInUser{}).Clauses(clause.Locking{Strength: "UPDATE"}).Where("appId = ?", 短链信息.AppId).Where("userId = ?", 短链信息.Uid).First(&info.checkInUser).Error
		if err != nil {
			return err
		}
		info.checkInUser.CheckInScore += info.CheckInInfo.ShareGivePoints
		_, err = service.NewCheckInUser(c, tx).UpdateMap([]int{info.checkInUser.Id}, map[string]interface{}{
			"checkInScore": info.checkInUser.CheckInScore,
		})

		//增加签到积分记录
		_, err = service.NewCheckInScoreLog(c, tx).Create(&dbm.DB_CheckInScoreLog{
			Id:           0,
			AppId:        短链信息.AppId,
			UserId:       短链信息.Uid,
			CreatedAt:    time.Now().Unix(),
			Number:       int64(info.CheckInInfo.ShareGivePoints),
			Msg:          "好友访问每日分享",
			NumberBefore: info.checkInUser.CheckInScore - info.CheckInInfo.ShareGivePoints,
			NumberAfter:  info.checkInUser.CheckInScore,
		})
		if err != nil {
			return err
		}
		return err
	})

	if err != nil {
		global.GVA_LOG.Error("签到分享被访问触发异常", zap.Any("err", err))
		return
	}

	return
}
