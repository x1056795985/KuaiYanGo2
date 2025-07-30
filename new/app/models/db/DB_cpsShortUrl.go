package db

// cps短链设计,
// 短连接唯一
// 长链接可重复,用户可能有多个短链接  不要删除,因为用户可能收藏短链,过几年在访问,如果删除,可能导致短链无法访问
type DB_CpsShortUrl struct {
	Id         int    `json:"id" gorm:"column:id;primarykey;"`
	CreatedAt  int64  `json:"createdAt" gorm:"column:createdAt;comment:创建时间"`
	UpdatedAt  int64  `json:"updatedAt" gorm:"column:updatedAt;;comment:更新时间"`
	ShortUrl   string `json:"shortUrl" gorm:"column:shortUrl;size:255;uniqueIndex;comment:短连接"`
	LongUrl    string `json:"longUrl" gorm:"column:longUrl;comment:长连接"`
	ClickCount int64  `json:"clickCount" gorm:"column:clickCount;default:0;comment:点击次数"`
	Uid        int    `json:"uid" gorm:"column:uid;comment:用户id"`
	Status     int    `json:"status" gorm:"column:status;default:1;comment:状态 1启用 0禁用"`
}

func (DB_CpsShortUrl) TableName() string {
	return "db_cps_short_url"
}
