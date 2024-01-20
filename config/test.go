package config

// 例子写出记录
type Test struct {
	DbAgentLevel     int `json:"db_agent_level"`
	DbAppinfo        int `json:"db_appinfo"`
	DbLogmoney       int `json:"db_logmoney"`
	DbLogrmbpayorder int `json:"db_logrmbpayorder"`
	DbLogusermsg     int `json:"db_logusermsg"`
	DbLogvipnumber   int `json:"db_logvipnumber"`
	DbPublicdata     int `json:"db_publicdata"`
	DbUser           int `json:"db_user"`
	Taskpool         int `json:"taskpool_类型"`
	User             int `json:"user"`
	Cron             int `json:"Cron"`
}
