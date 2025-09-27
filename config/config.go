package config

import m "server/new/app/models/common"

type Server struct {
	AgentUid int     `mapstructure:"duid" json:"duid"`        //代理uid
	Port     int     `mapstructure:"Port" json:"Port" `       // 端口值
	Zap      Zap     `mapstructure:"zap" json:"zap" `         //日志配置
	Captcha  Captcha `mapstructure:"captcha" json:"captcha" ` //验证码配置
	Q取运行目录   string
	// gorm
	Mysql Mysql `mapstructure:"mysql" json:"mysql" `
}
type Server备用 struct {
	Z在线支付      m.Z在线支付    `mapstructure:"在线支付" json:"在线支付" `           // 系统名称
	D短信平台配置    D短信平台配置    `mapstructure:"短信平台配置" json:"短信平台配置" `       // 短信配置
	X行为验证码平台配置 X行为验证码平台配置 `mapstructure:"行为验证码平台配置" json:"行为验证码平台配置" ` // 短信配置
	Captcha    Captcha    `mapstructure:"captcha" json:"captcha" `     //验证码配置
}

type X系统设置 struct {
	X系统名称      string `mapstructure:"系统名称" json:"系统名称" ` // 系统名称
	X系统地址      string `mapstructure:"系统地址" json:"系统地址" `
	Y用户API加密盐  string `mapstructure:"用户API加密盐" json:"用户API加密盐" `
	G管理员后台Host string `mapstructure:"管理员后台Host" json:"管理员后台Host" `
	WebApiHost string `mapstructure:"WebApiHost" json:"WebApiHost" `
	D代理后台Host  string `mapstructure:"代理后台Host" json:"代理后台Host" `
	X系统开关      bool   `mapstructure:"系统开关" json:"系统开关" `
	X系统关闭提示    string `mapstructure:"系统关闭提示" json:"系统关闭提示" `
	Y用户中心开关    bool   `mapstructure:"用户中心开关" json:"用户中心开关" `
	D代理中心开关    bool   `mapstructure:"代理中心开关" json:"代理中心开关" `
	D代理中心关闭提示  string `mapstructure:"代理中心关闭提示" json:"代理中心关闭提示" `
	//W系统模式      int    `mapstructure:"系统模式" json:"系统模式" ` // 0  正常用户 1 演示模式     1056795985 调试模式
	B备案号 string `mapstructure:"备案号" json:"备案号" `
}

type D短信平台配置 struct {
	D当前选择    int      `mapstructure:"当前选择" json:"当前选择" `
	TX云短信Sms TX云短信Sms `mapstructure:"TX云Sms" json:"TX云Sms" `
	Sms短信宝   Sms短信宝   `mapstructure:"Sms短信宝" json:"Sms短信宝" `
	Sms七牛云   Sms七牛云   `mapstructure:"Sms七牛云" json:"Sms七牛云" `
}

//id:1
type TX云短信Sms struct {
	SECRET_ID  string `mapstructure:"SECRET_ID" json:"SECRET_ID" `
	SECRET_KEY string `mapstructure:"SECRET_KEY" json:"SECRET_KEY" `
	D短信应用ID    string `mapstructure:"短信应用ID" json:"短信应用ID" `
	D短信签名      string `mapstructure:"短信签名" json:"短信签名"`
	Z正文模板ID    string `mapstructure:"正文模板ID" json:"正文模板ID" `
}

// Id:2
type Sms短信宝 struct {
	User   string `mapstructure:"User" json:"User" `
	ApiKey string `mapstructure:"ApiKey" json:"ApiKey" `
	C产品Id  string `mapstructure:"ProductId" json:"ProductId" `
	F发送内容  string `mapstructure:"SendValue" json:"SendValue" `
}

// Id:3
type Sms七牛云 struct {
	AccessKey   string `json:"AccessKey" `
	SecretKey   string `json:"SecretKey" `
	SignatureID string `json:"SignatureID" ` //签名id
	TemplateID  string `json:"TemplateID" `  //模板id
}

type X行为验证码平台配置 struct {
	D当前选择    int     `mapstructure:"当前选择" json:"当前选择" `
	J极验行为验证4 极验行为验证4 `mapstructure:"极验行为验证4" json:"极验行为验证4" `
}

type 极验行为验证4 struct {
	Y验证_ID  string `mapstructure:"验证_ID" json:"验证_ID" `
	Y验证_KEY string `mapstructure:"验证_KEY" json:"验证_KEY" `
}

type MQTT配置 struct {
	L连接状态  bool   `json:"连接状态"`
	F服务器地址 string `json:"服务器地址"`
	F服务器端口 int    `json:"服务器端口"`
	Y用户名   string `json:"用户名"`
	M密码    string `json:"密码"`
}
type Y云存储配置 struct {
	D当前选择    int      `mapstructure:"当前选择" json:"当前选择" ` // 1 S3通用协议 2 七牛云
	S3兼容协议   S3兼容协议   `mapstructure:"S3兼容协议" json:"S3兼容协议" `
	Q七牛云对象存储 Q七牛云对象存储 `mapstructure:"七牛云对象存储" json:"七牛云对象存储" `
}
type Q七牛云对象存储 struct {
	AccessKey string `  json:"AccessKey" `
	SecretKey string `  json:"SecretKey" `
	W外链域名     string `json:"外链域名" `       //外链域名
	Bucket    string `  json:"Bucket" `   //空间名称
	RootPath  string `  json:"rootPath" ` //根文件夹
}
type S3兼容协议 struct {
	Endpoint  string `  json:"Endpoint" `
	AccessKey string `  json:"AccessKey" `
	SecretKey string `  json:"SecretKey" `
	W外链域名     string `json:"外链域名" `       //外链域名
	Bucket    string `  json:"Bucket" `   //空间名称
	RootPath  string `  json:"rootPath" ` //根文件夹
}
type Y用户消息配置 struct {
	MsgTypeList string `  json:"MsgTypeList" `
}
