package common

// 文件详情
type W文件对象详情 struct {
	Name   string `json:"Name"`   // 对象名称
	Path   string `json:"Path"`   // 对象路径
	Type   int    `json:"Type"`   // 对象类型  1 文件 2 文件夹
	UpTime int64  `json:"UpTime"` // 上传时间戳
	Size   int64  `json:"Size"`   // 对象大小，单位为字节
	MD5    string `json:"MD5"`    // 对象 MD5 值
}

// 上传凭证
type W文件上传凭证 struct {
	Path    string `json:"Path"`    // 对象路径
	Type    int    `json:"Type"`    // 对象类型  1:自身 2:七牛云
	Url     string `json:"Url"`     // 上传url
	UpToken string `json:"UpToken"` // UpToken
}
