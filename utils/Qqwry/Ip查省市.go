package Qqwry

import (
	_ "embed"
)

// 查询次数较少,为了方便,还是每次打开查询关闭数据库,后去如果需要查询性能,在处理
func C查询IP归属地(ip地址 string) (string, error) {
	//局_耗时 := time.Now().UnixMilli()
	省份, _, err := QueryIP(ip地址)
	if err != nil {
		return "", err
	}
	//fmt.Printf("ip查询耗时:%dms,城市：%s，运营商：%s\n", time.Now().UnixMilli()-局_耗时, city, isp)
	return 省份, nil
}
