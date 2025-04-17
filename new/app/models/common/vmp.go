package common

type VmpParams struct {
	UserName     string `json:"UserName,omitempty"`  //用户名
	Email        string `json:"Email,omitempty"`     //邮箱地址
	Hwid         string `json:"Hwid,omitempty"`      //机器码
	ExpireDate   S时间    `json:"ExpireDate"`          //到期时间  年月日有效  时分秒无效
	MaxBuildDate S时间    `json:"MaxBuildDate"`        //最大创建时间
	TimeLimit    int    `json:"TimeLimit,omitempty"` //运行次数
	UserData     []byte `json:"UserData,omitempty"`  //用户数据
}
type S时间 struct {
	Year  int `json:"Year,omitempty"`
	Month int `json:"Month,omitempty"`
	Day   int `json:"Day,omitempty"`
}

// Rsa位数 int, RsaBase64私钥, RsaBase64模数, base64产品代码 string
type VmpRsa struct {
	Rsa位数       int    `json:"Rsa位数,omitempty"`
	RsaBase64私钥 string `json:"RsaBase64私钥,omitempty"`
	RsaBase64模数 string `json:"RsaBase64模数,omitempty"`
	Base64产品代码  string `json:"Base64产品代码,omitempty"`
}
