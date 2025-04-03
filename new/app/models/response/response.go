package response

import dbm "server/new/app/models/db"

type GetList struct {
	List  interface{} `json:"List"`  // 列表
	Count int64       `json:"Count"` // 总数
}

type KaClassUp带调价信息 struct {
	dbm.DB_KaClass         //卡类信息
	UserClassName  string  `json:"UserClassName"` //用户类型名称
	Markup         float64 `json:"Markup"`        //调整价
}
