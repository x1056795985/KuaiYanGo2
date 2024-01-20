package request

//请求列表通用结构体

// 列表请求通用参数
type List struct {
	Page     int    `json:"Page" binding:"required"` // 页
	Size     int    `json:"Size" binding:"required"` // 页数量
	Type     int    `json:"Type"`                    // 关键字类型
	Keywords string `json:"Keywords"`                // 关键字
	Order    int    `json:"Order"`                   // 0 倒序 1 正序
	Count    int64  `json:"Count"`                   // 总数缓存
}

// 单id请求
type Id struct {
	Id int `json:"Id" binding:"required,min=1"`
}

// 单id数组请求
type Ids struct {
	Ids []int `json:"Ids" binding:"required,min=1"` //id数组
}
