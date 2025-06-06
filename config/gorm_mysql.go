package config

type Mysql struct {
	Path         string `mapstructure:"Path" json:"Path" `                      // 服务器地址
	Port         string `mapstructure:"Port" json:"Port" `                      //:端口
	Config       string `mapstructure:"Config" json:"Config" `                  // 高级配置
	Dbname       string `mapstructure:"Dbname" json:"Dbname" `                  // 数据库名
	Username     string `mapstructure:"Username" json:"Username" `              // 数据库用户名
	Password     string `mapstructure:"Password" json:"Password" `              // 数据库密码
	Prefix       string `mapstructure:"Prefix" json:"Prefix" `                  //全局表前缀，单独定义TableName则不生效
	Singular     bool   `mapstructure:"Singular" json:"Singular" `              //是否开启全局禁用复数，true表示开启
	Engine       string `mapstructure:"Engine" json:"Engine"  default:"InnoDB"` //数据库引擎，默认InnoDB
	MaxIdleConns int    `mapstructure:"MaxIdleConns" json:"MaxIdleConns" `      // 空闲中的最大连接数
	MaxOpenConns int    `mapstructure:"MaxOpenConns" json:"MaxOpenConns" `      // 打开到数据库的最大连接数
	LogMode      string `mapstructure:"LogMode" json:"LogMode" `                // 是否开启Gorm全局日志
	LogZap       bool   `mapstructure:"LogZap" json:"LogZap" `                  // 是否通过zap写入日志文件
}

// 链接dsn 字符串
func (m *Mysql) Dsn() string {
	return m.Username + ":" + m.Password + "@tcp(" + m.Path + ":" + m.Port + ")/" + m.Dbname + "?" + m.Config
}

// 返回 日志类型
func (m *Mysql) GetLogMode() string {
	return m.LogMode
}
