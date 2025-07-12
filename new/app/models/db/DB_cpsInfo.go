package db

// cps推广活动配置表
type CpsInfo struct {
	Id                 int   `json:"id" gorm:"column:id;primarykey;autoIncrement:false;comment:关联活动表id"`
	CreateTime         int64 `json:"createTime" gorm:"column:createTime;comment:创建时间戳"`
	UpdateTime         int64 `json:"updateTime" gorm:"column:updateTime;comment:更新时间戳"`
	BronzeThreshold    int   `json:"bronzeThreshold" gorm:"column:bronzeThreshold;comment:成为铜牌推广数量阈值"` //虽然铜牌居然还需要有数量限制,但是确实是这样的,比如快快网络,前两个推广是不给佣金的,说明有需求,可能是防止自己开小号成为自己的推广者
	BronzeKickback     int   `json:"bronzeKickback" gorm:"column:bronzeKickback;comment:铜牌分成比例"`
	SilverThreshold    int   `json:"silverThreshold" gorm:"column:silverThreshold;comment:成为银牌推广数量阈值"`
	SilverKickback     int   `json:"silverKickback" gorm:"column:silverKickback;comment:银牌分成比例"`
	GoldMedalThreshold int   `json:"goldMedalThreshold" gorm:"column:goldMedalThreshold;comment:成为金牌推广数量阈值"`
	GoldMedalKickback  int   `json:"goldMedalKickback" gorm:"column:goldMedalKickback;comment:金牌分成比例"`
	GrandsonKickback   int   `json:"grandsonKickback" gorm:"column:grandsonKickback;comment:徒孙分成比例"` //让推广者愿意教别人怎么推广
	NarrowPic          int   `json:"widePic" gorm:"column:widePic;comment:素材_窄图"`
	DetailPic          int   `json:"detailPic" gorm:"column:detailPic;comment:素材_详情图"`
}

func (CpsInfo) TableName() string {
	return "db_cps_info"
}
