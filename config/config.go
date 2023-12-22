package config

type Server struct {
	Port    int     `mapstructure:"Port" json:"Port" yaml:"Port"`          // 端口值
	Zap     Zap     `mapstructure:"zap" json:"zap" yaml:"zap"`             //日志配置
	Captcha Captcha `mapstructure:"captcha" json:"captcha" yaml:"captcha"` //验证码配置
	Q取运行目录  string
	// gorm
	Mysql Mysql `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
}
type Server备用 struct {
	X系统设置      X系统设置      `mapstructure:"系统设置" json:"系统设置" yaml:"系统设置"`                // 系统名称
	Z在线支付      Z在线支付      `mapstructure:"在线支付" json:"在线支付" yaml:"在线支付"`                // 系统名称
	D短信平台配置    D短信平台配置    `mapstructure:"短信平台配置" json:"短信平台配置" yaml:"短信平台配置"`          // 短信配置
	X行为验证码平台配置 X行为验证码平台配置 `mapstructure:"行为验证码平台配置" json:"行为验证码平台配置" yaml:"行为验证码平台配置"` // 短信配置
	Captcha    Captcha    `mapstructure:"captcha" json:"captcha" yaml:"captcha"`       //验证码配置
}

type X系统设置 struct {
	X系统名称      string `mapstructure:"系统名称" json:"系统名称" yaml:"系统名称"` // 系统名称
	X系统地址      string `mapstructure:"系统地址" json:"系统地址" yaml:"系统地址"`
	Y用户API加密盐  string `mapstructure:"用户API加密盐" json:"用户API加密盐" yaml:"用户API加密盐"`
	G管理员后台Host string `mapstructure:"管理员后台Host" json:"管理员后台Host" yaml:"管理员后台Host"`
	WebApiHost string `mapstructure:"WebApiHost" json:"WebApiHost" yaml:"WebApiHost"`
	D代理后台Host  string `mapstructure:"代理后台Host" json:"代理后台Host" yaml:"代理后台Host"`
	X系统开关      bool   `mapstructure:"系统开关" json:"系统开关" yaml:"系统开关"`
	X系统关闭提示    string `mapstructure:"系统关闭提示" json:"系统关闭提示" yaml:"系统关闭提示"`
	Y用户中心开关    bool   `mapstructure:"用户中心开关" json:"用户中心开关" yaml:"用户中心开关"`
	D代理中心开关    bool   `mapstructure:"代理中心开关" json:"代理中心开关" yaml:"代理中心开关"`
	D代理中心关闭提示  string `mapstructure:"代理中心关闭提示" json:"代理中心关闭提示" yaml:"代理中心关闭提示"`
	//W系统模式      int    `mapstructure:"系统模式" json:"系统模式" yaml:"系统模式"` // 0  正常用户 1 演示模式     1056795985 调试模式
	B备案号 string `mapstructure:"备案号" json:"备案号" yaml:"备案号"`
}

type Z在线支付 struct {
	J禁止退款 bool `mapstructure:"禁止退款" json:"禁止退款" yaml:"禁止退款"`

	Z支付宝开关      bool   `mapstructure:"支付宝开关" json:"支付宝开关" yaml:"支付宝开关"`
	Z支付宝显示名称    string `mapstructure:"支付宝显示名称" json:"支付宝显示名称"`
	Z支付宝商户ID    string `mapstructure:"支付宝商户ID" json:"支付宝商户ID" yaml:"支付宝商户ID"`
	Z支付宝商户私钥    string `mapstructure:"支付宝商户私钥" json:"支付宝商户私钥" yaml:"支付宝商户私钥"`
	Z支付宝商户公钥    string `mapstructure:"支付宝商户公钥" json:"支付宝商户公钥" yaml:"支付宝商户公钥"`
	Z支付宝公钥      string `mapstructure:"支付宝公钥" json:"支付宝公钥" yaml:"支付宝公钥"`
	Z支付宝同步回调url string `mapstructure:"支付宝同步回调url" json:"支付宝同步回调url" yaml:"支付宝同步回调url"`
	Z支付宝单次最大金额  int    `mapstructure:"支付宝单次最大金额" json:"支付宝单次最大金额" yaml:"支付宝单次最大金额"`

	Z支付宝当面付开关      bool   `mapstructure:"支付宝当面付开关" json:"支付宝当面付开关" yaml:"支付宝当面付开关"`
	Z支付宝当面付显示名称    string `mapstructure:"支付宝当面付显示名称" json:"支付宝当面付显示名称"`
	Z支付宝当面付商户ID    string `mapstructure:"支付宝当面付商户ID" json:"支付宝当面付商户ID" yaml:"支付宝当面付商户ID"`
	Z支付宝当面付商户私钥    string `mapstructure:"支付宝当面付商户私钥" json:"支付宝当面付商户私钥" yaml:"支付宝当面付商户私钥"`
	Z支付宝当面付商户公钥    string `mapstructure:"支付宝当面付商户公钥" json:"支付宝当面付商户公钥" yaml:"支付宝当面付商户公钥"`
	Z支付宝当面付公钥      string `mapstructure:"支付宝当面付公钥" json:"支付宝当面付公钥" yaml:"支付宝当面付公钥"`
	Z支付宝当面付同步回调url string `mapstructure:"支付宝当面付同步回调url" json:"支付宝当面付同步回调url" yaml:"支付宝当面付同步回调url"`
	Z支付宝当面付单次最大金额  int    `mapstructure:"支付宝当面付单次最大金额" json:"支付宝当面付单次最大金额" yaml:"支付宝当面付单次最大金额"`

	Z支付宝H5开关      bool   `mapstructure:"支付宝H5开关" json:"支付宝H5开关"`
	Z支付宝H5显示名称    string `mapstructure:"支付宝H5显示名称" json:"支付宝H5显示名称"`
	Z支付宝H5商户ID    string `mapstructure:"支付宝H5商户ID" json:"支付宝H5商户ID" `
	Z支付宝H5商户私钥    string `mapstructure:"支付宝H5商户私钥" json:"支付宝H5商户私钥" `
	Z支付宝H5商户公钥    string `mapstructure:"支付宝H5商户公钥" json:"支付宝H5商户公钥"`
	Z支付宝H5公钥      string `mapstructure:"支付宝H5公钥" json:"支付宝H5公钥" `
	Z支付宝H5同步回调url string `mapstructure:"支付宝H5同步回调url" json:"支付宝H5同步回调url"`
	Z支付宝H5单次最大金额  int    `mapstructure:"支付宝H5单次最大金额" json:"支付宝H5单次最大金额" `

	W微信支付开关      bool   `mapstructure:"微信支付开关" json:"微信支付开关" yaml:"微信支付开关"`
	W微信支付显示名称    string `mapstructure:"微信支付显示名称" json:"微信支付显示名称"`
	W微信支付商户ID    string `mapstructure:"微信支付商户ID" json:"微信支付商户ID" yaml:"微信支付商户ID"`
	W微信支付AppId   string `mapstructure:"微信支付AppId" json:"微信支付AppId" yaml:"微信支付AppId"`
	W微信支付商户v3密钥  string `mapstructure:"微信支付商户v3密钥" json:"微信支付商户v3密钥" yaml:"微信支付商户v3密钥"`
	W微信支付商户证书串   string `mapstructure:"微信支付商户证书串" json:"微信支付商户证书串" yaml:"微信支付商户证书串"`
	W微信支付商户证书序列号 string `mapstructure:"微信支付商户证书序列号" json:"微信支付商户证书序列号" yaml:"微信支付商户证书序列号"`
	W微信支付异步回调Url string `mapstructure:"微信支付异步回调Url" json:"微信支付异步回调Url" yaml:"微信支付异步回调Url"`
	W微信支付单次最大金额  int    `mapstructure:"微信支付单次最大金额" json:"微信支付单次最大金额" yaml:"微信支付单次最大金额"`

	X小叮当支付开关   bool   `mapstructure:"小叮当支付开关" json:"小叮当支付开关"`
	X小叮当支付显示名称 string `mapstructure:"小叮当支付显示名称" json:"小叮当支付显示名称"`
	X小叮当app_id string `mapstructure:"小叮当app_id" json:"小叮当app_id"`
	X小叮当接口密钥   string `mapstructure:"小叮当接口密钥" json:"小叮当接口密钥" `
	X小叮当支付类型   int    `mapstructure:"小叮当支付类型" json:"小叮当支付类型" `
	X小叮当单次最大金额 int    `mapstructure:"小叮当单次最大金额" json:"小叮当单次最大金额" `
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
