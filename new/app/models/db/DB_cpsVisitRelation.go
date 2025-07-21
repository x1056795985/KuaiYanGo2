package db

// cps关系表
type DB_CpsVisitRelation struct {
	Id           int   `json:"id" gorm:"column:id;primarykey;comment:关系ID"`
	CreatedAt    int64 `json:"createdAt" gorm:"column:createdAt;comment:创建时间"`
	UpdatedAt    int64 `json:"updatedAt" gorm:"column:updatedAt;comment:更新时间"`
	VisitUserId  int   `json:"visitUserId" gorm:"column:visitUserId;index;comment:邀请者用户ID"` //邀请者一定是账号id   邀请者是余额结算,无需AppId
	VisitedAppId int   `json:"visitedAppId" gorm:"column:visitedAppId;index:idx_appid_userid;comment:被邀请者AppId"`
	//不同应用可能是不同的邀请者,要区分开来, 用户在该应用充值时,给对应的邀请者
	//还有种情况,账号模式应用 a 第一次邀请的c注册应用1,这时候,b就无法邀请注册了,但是又确实是b邀请的注册应用2,只能让C登陆后,手动填写邀请人
	VisitedUserId int    `json:"visitedUserId" gorm:"column:visitedUserId;index:idx_appid_userid;comment:被邀请者用户或卡号ID"` //被邀请者可能是账号,可能是卡号,要根据appid分别处理
	Level         int    `json:"level" gorm:"column:level;comment:邀请等级 1-一级 2-二级等"`                                    //假设我们做了两级分销。当前注册的用户C的邀请者是B，然后去查邀请表有A推荐B注册的，那么我们就可以在邀请关系表里面新增两条记录
	Status        int    `json:"status" gorm:"column:status;default:1;comment:状态 1-正常 0-失效"`                           //单独设置一个,不用每次都比对时间,影响数据库性能
	Referer       string `json:"referer" gorm:"column:referer;comment:来源"`                                             //来源 从哪里注册推广的,方便统计 类似广告位id
}

func (DB_CpsVisitRelation) TableName() string {
	return "db_cps_visit_relation"
}
