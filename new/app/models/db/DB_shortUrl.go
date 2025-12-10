package db

// 短链设计,
// 短连接唯一
// 长链接可重复,用户可能有多个短链接  不要删除,因为用户可能收藏短链,过几年在访问,如果删除,可能导致短链无法访问
type DB_ShortUrl struct {
	Id         int    `json:"id" gorm:"column:id;primarykey;AUTO_INCREMENT;"`
	CreatedAt  int64  `json:"createdAt" gorm:"column:createdAt;comment:创建时间"`
	UpdatedAt  int64  `json:"updatedAt" gorm:"column:updatedAt;;comment:更新时间"`
	ShortUrl   string `json:"shortUrl" gorm:"column:shortUrl;size:255;uniqueIndex;comment:短连接"` //用于区分其他业务的短链
	BaseUrl    string `json:"baseUrl" gorm:"column:baseUrl;comment:基础地址"`                       //用于跳转本地跳转页,
	RouterUrl  string `json:"routerUrl" gorm:"column:routerUrl;comment:跳转路由地址"`                 //因为是基于vue(含小程序)的所以不能只用一个长链接
	ClickCount int64  `json:"clickCount" gorm:"column:clickCount;default:0;comment:点击次数"`
	Uid        int    `json:"uid" gorm:"column:uid;comment:用户id"`
	AppId      int    `json:"appId" gorm:"column:appId;comment:来源appId"`
	Status     int    `json:"status" gorm:"column:status;default:1;comment:状态 1启用 2禁用"`
	ShortType  int    `json:"shortType" gorm:"column:shortType;default:0;comment:短链类型,关系到访问后处理,1签到生成"`
	Other      string `json:"other" gorm:"column:other;type:varchar(5000);comment:其他"`
}

func (DB_ShortUrl) TableName() string {
	return "db_short_url"
}
