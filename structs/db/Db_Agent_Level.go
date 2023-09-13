package DB

// https://blog.csdn.net/qq_25062671/article/details/108653033  结构参考文章
type Db_Agent_Level struct {
	ID        int `json:"Id" gorm:"column:Id;primaryKey;comment:Id"`
	Uid       int `json:"Uid" gorm:"column:Uid;comment:用户Id"`
	UPAgentId int `json:"UPAgentId" gorm:"column:UPAgentId;comment:上级代理ID"` //负数为管理员ID,一般只有一级代理上级代理才是负数
	Level     int `json:"Level" gorm:"column:Level;comment:是上级代理的第几级代理"`
}

/*
一级代理 占一条记录, 二级代理占两条记录,三级代理占三条记录,
userid(-1)->userid(2)->userid(4)
userid(-1)->userid(3)
userid(-1)->userid(5)->userid(6)->userid(7)
管理员     ->一级代理  ->二级代理  -> 三级代理


表结构
用户ID | 上级代理ID  |是上级代理ID的几级代理
2	  |  -1			| 1
3	  |  -1			| 1
4	  |  2			| 1
4	  |  -1			| 2

5	  |  -1			| 1

6	  |  5			| 1
6	  |  -1			| 2

7	  |  6			| 1
7	  |  5			| 2
7	  |  -1			| 3
*/

func (Db_Agent_Level) TableName() string {
	return "db_Agent_Level" //代理关系等级表
}
