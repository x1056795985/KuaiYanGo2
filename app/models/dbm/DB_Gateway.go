package dbm

// DB_Gateway 网关转发配置表,每个网关对应一个业务服务
// 网关功能:客户端带 Token+WangGuanId 请求 /wangGuan/*,鉴权后签发JWT并透明转发到 Url
type DB_Gateway struct {
	Id         int    `json:"Id" gorm:"column:Id;primarykey;AUTO_INCREMENT"`                        // =WangGuanId
	Name       string `json:"Name" gorm:"column:Name;size:100;comment:网关名称"`                        // 名称
	Url        string `json:"Url" gorm:"column:Url;size:255;comment:业务服务地址,如http://127.0.0.1:18899"` // 业务服务根地址
	Secret     string `json:"Secret" gorm:"column:Secret;size:100;comment:JWT签发密钥(每个网关独立,严禁共用)"`     // 签发密钥
	Status     int    `json:"Status" gorm:"column:Status;default:1;comment:1启用 2禁用"`                 // 状态
	TimeoutSec int    `json:"TimeoutSec" gorm:"column:TimeoutSec;default:60;comment:业务响应头超时秒数"`       // 超时
	Remark     string `json:"Remark" gorm:"column:Remark;size:255;comment:备注"`                       // 备注
}

func (DB_Gateway) TableName() string {
	return "db_Gateway"
}
