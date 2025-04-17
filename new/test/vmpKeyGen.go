package main

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"server/new/app/models/common"
	"time"
)

func test() {
	// 配置参数（示例值）
	params := VmpParams{
		UserName:     "abc",
		Email:        "abc@qq.com",
		Hwid:         "wO1Xs+VL+7afzVpicvaT/YrMrwdO/aBLusH8eg==",
		ExpireDate:   common.S时间{2028, 4, 14}, //有效但是只到天
		MaxBuildDate: common.S时间{2025, 4, 15},
		TimeLimit:    1,
		UserData:     []byte{},
	}

	exported := struct {
		Algorithm   string
		Bits        int
		PrivateKey  string
		Modulus     string
		ProductCode string
	}{
		Algorithm:   "RSA",
		Bits:        1024,
		PrivateKey:  "hjJi7NNcxxZ3Me3syJiRamoCK0kuXtun4JMctTTavf895giOXzMXKaRcW3MxtKGPhT1bAuY8wgMRrfgeNrS6/eJBiK3n06jWy2g04Kcnp5Q/rptrd6YG9j+vcBmScNgaseUQmkT5gc6ujlLo0P3W4oWkf4mJZNJKO0t6+6QEHEE=",
		Modulus:     "5HSlarnEogn6B/KIOCkVPVPx9s45M9KWFs1lePOY/8szCG2sHe8jalkihKlyQ3b15BAlxeAJ1+0+zTau3tSAUGTJs6s+f2AtLOYWFoyBv1PnNMh2tyDOdIrLYz11VjgN3igD1r6fMA6kpbm2SwUt5EL9MQDpkqfbBSdcBF0NWzE=",
		ProductCode: "AAAAAQAAJxE=",
	}
	var 计算授权码 string
	局_耗时 := time.Now().Unix()
	//for _ = range 10000 {
	计算授权码, _ = J计算授权码(exported.Bits, exported.PrivateKey, exported.Modulus, exported.ProductCode, params)
	//}
	局_耗时 = time.Now().Unix() - 局_耗时
	fmt.Printf("耗时：%d ms\n", 局_耗时) //耗时：7 ms  签名还是很快的,万次才7毫秒

	fmt.Println(计算授权码)
}

func J计算授权码(Rsa位数 int, RsaBase64私钥, RsaBase64模数, base64产品代码 string, 授权信息 VmpParams) (授权码 string, err error) {
	exported := struct {
		Algorithm   string
		Bits        int
		PrivateKey  string
		Modulus     string
		ProductCode string
	}{
		Algorithm:   "RSA",
		Bits:        Rsa位数,
		PrivateKey:  RsaBase64私钥,
		Modulus:     RsaBase64模数,
		ProductCode: base64产品代码,
	}

	serialData, err2 := 打包数据(授权信息, exported.ProductCode)
	if err2 != nil {
		err = err2
		return
	}
	serialBin := serialData

	hash := sha1.Sum(serialBin)
	serialBin = append(serialBin, 0xFF)
	for i := 3; i >= 0; i-- {
		serialBin = append(serialBin, hash[i])
	}

	// 生成 8 到 12 之间的随机整数
	//$size = rand(8, 16);
	//for ($i = 0; $i < $size; $i++) $padding_front[] = rand(1, 255);
	// 改为固定值（示例用全0填充）
	paddingFront := make([]byte, 8)
	for i := range paddingFront {
		paddingFront[i] = byte(rand.Int31n(255) + 1)
	}

	paddingFront = append([]byte{0, 2}, paddingFront...)
	paddingFront = append(paddingFront, 0)

	contentSize := len(serialBin) + len(paddingFront)
	rest := exported.Bits/8 - contentSize
	if rest < 0 {
		err = errors.New("内容太大")
		return
	}

	paddingBack := make([]byte, rest)
	finalSerial := append(paddingFront, serialBin...)
	finalSerial = append(finalSerial, paddingBack...)

	n, _ := base64.StdEncoding.DecodeString(exported.Modulus)
	d, _ := base64.StdEncoding.DecodeString(exported.PrivateKey)

	encoded := 十进制编码(finalSerial)
	modulus := 十进制编码(n)
	private := 十进制编码(d)

	res := new(big.Int).Exp(encoded, private, modulus) //	 模幂运算
	resBytes := 十进制解码(res)
	//fmt.Println(string(resBytes))
	授权码 = base64.StdEncoding.EncodeToString(resBytes)

	return
}

// 辅助函数
func 十进制编码(data []byte) *big.Int {
	result := big.NewInt(0)
	for _, b := range data {
		result.Mul(result, big.NewInt(256))
		result.Add(result, big.NewInt(int64(b)))
	}
	return result
}

func 十进制解码(n *big.Int) []byte {
	result := []byte{}
	zero := big.NewInt(0)
	twoFiftySix := big.NewInt(256)

	for n.Cmp(zero) > 0 {
		mod := new(big.Int)
		mod.Mod(n, twoFiftySix)
		result = append([]byte{byte(mod.Int64())}, result...)
		n.Div(n, twoFiftySix)
	}
	return result
}

// 主要逻辑
type VmpParams struct {
	UserName     string     //用户名
	Email        string     //邮箱地址
	Hwid         string     //机器码
	ExpireDate   common.S时间 //到期时间  年月日有效  时分秒无效
	MaxBuildDate common.S时间 //最大创建时间
	TimeLimit    int        //运行次数
	UserData     []byte     //用户数据
}

func 打包数据(params VmpParams, productCode string) (结果数据 []byte, err error) {
	serial := []byte{1, 1} // Version
	// 用户名
	if len(params.UserName) > 255 {
		err = errors.New("用户名太长")
		return
	}
	serial = append(serial, 2, byte(len(params.UserName)))
	serial = append(serial, []byte(params.UserName)...)

	// Email
	if len(params.Email) > 255 {
		err = errors.New("E-Mail 太长")
		return
	}
	serial = append(serial, 3, byte(len(params.Email)))
	serial = append(serial, []byte(params.Email)...)

	// Hardware ID
	hwid, _ := base64.StdEncoding.DecodeString(params.Hwid)
	if len(hwid) == 0 || len(hwid) > 255 || len(hwid)%4 != 0 {
		err = errors.New("机器码 无效")
		return
	}
	serial = append(serial, 4, byte(len(hwid)))
	serial = append(serial, hwid...)

	// Expire Date
	serial = append(serial, 5)
	serial = appendDate(serial, params.ExpireDate)

	// Time Limit
	if params.TimeLimit < 0 || params.TimeLimit > 255 {
		err = errors.New("无效的时限")
		return
	}
	serial = append(serial, 6, byte(params.TimeLimit))

	// Product Code
	pc, _ := base64.StdEncoding.DecodeString(productCode)
	if len(pc) != 8 {
		err = errors.New("产品代码无效")
		return
	}
	serial = append(serial, 7)
	serial = append(serial, pc...)

	// User Data
	if len(params.UserData) > 255 {
		err = errors.New("用户数据过长")
		return
	}
	serial = append(serial, 8, byte(len(params.UserData)))
	serial = append(serial, params.UserData...)

	// Max Build Date
	serial = append(serial, 9)
	serial = appendDate(serial, params.MaxBuildDate)
	结果数据 = serial
	return
}

func appendDate(serial []byte, a common.S时间) []byte {
	var t time.Time
	t = time.Date(a.Year, time.Month(a.Month), a.Day, 24, 59, 59, 0, time.UTC)
	year, month, day := t.Date()
	return append(serial, byte(day), byte(month), byte(year%256), byte(year/256))

}
