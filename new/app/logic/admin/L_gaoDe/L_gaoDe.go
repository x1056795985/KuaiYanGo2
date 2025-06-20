package L_gaoDe

import (
	. "EFunc/utils"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/imroc/req/v3"
	"github.com/valyala/fastjson"
	"server/global"
)

func G高德查询天气(c *gin.Context) (data2 string, err error) {
	key := ""
	if global.Q快验.Q取应用专属变量(&key, "高德key") {
		key = fastjson.GetString([]byte(key), "高德key")
	} else {
		err = errors.New(global.Q快验.Q取错误信息(nil))
		return
	}
	if key == "" {
		err = errors.New(global.Q快验.Q取错误信息(nil))
		return
	}

	result, err2 := req.C().EnableInsecureSkipVerify().R().Get("https://restapi.amap.com/v3/ip?ip=" + S三元(c.ClientIP() == "127.0.0.1", global.Q快验.Q取用户IP(), c.ClientIP()) + "&key=" + key)
	if err2 != nil {
		return "", err2
	}
	data2 = fastjson.GetString(result.Bytes(), "adcode")

	result, err2 = req.C().EnableInsecureSkipVerify().R().Get("https://restapi.amap.com/v3/weather/weatherInfo?key=" + key + "&extensions=base&city=" + data2)
	if err2 != nil {
		return "", err2
	}
	/*	if (response.data.status === '1') {
		const s = response.data.lives[0]
		weatherInfo.value = s.city + ' 天气：' + s.weather + ' 温度：' + s.temperature + '摄氏度 风向：' + s.winddirection + ' 风力：' + s.windpower + '级 空气湿度：' + s.humidity
	}*/
	if "1" == fastjson.GetString(result.Bytes(), "status") {
		data2 = ""
		data2 += fastjson.GetString(result.Bytes(), "lives", "0", "city")
		data2 += " 天气："
		data2 += fastjson.GetString(result.Bytes(), "lives", "0", "weather")
		data2 += " 温度："
		data2 += fastjson.GetString(result.Bytes(), "lives", "0", "temperature")
		data2 += "摄氏度 风向："
		data2 += fastjson.GetString(result.Bytes(), "lives", "0", "winddirection")
		data2 += " 风力："
		data2 += fastjson.GetString(result.Bytes(), "lives", "0", "windpower")
		data2 += "级 空气湿度："
		data2 += fastjson.GetString(result.Bytes(), "lives", "0", "humidity")
	} else {
		err = errors.New(fastjson.GetString(result.Bytes(), "info"))
	}
	return
}

func X心知天气查询天气(c *gin.Context) (data2 string, err error) {
	key := ""
	if global.Q快验.Q取应用专属变量(&key, "心知天气") {
		key = fastjson.GetString([]byte(key), "心知天气")
	} else {
		err = errors.New(global.Q快验.Q取错误信息(nil))
		return
	}
	if key == "" {
		err = errors.New(global.Q快验.Q取错误信息(nil))
		return
	}

	result, err2 := req.C().EnableInsecureSkipVerify().R().Get("https://api.seniverse.com/v3/weather/now.json?key=S62cg9ezY6vo_zdI2&location=" + S三元(c.ClientIP() == "127.0.0.1", global.Q快验.Q取用户IP(), c.ClientIP()) + "&language=zh-Hans&unit=c")
	if err2 != nil {
		return "", err2
	}
	//{"results":[{"location":{"id":"WWYMRT0VRMUG","name":"大连","country":"CN","path":"大连,大连,辽宁,中国","timezone":"Asia/Shanghai","timezone_offset":"+08:00"},"now":{"text":"晴","code":"0","temperature":"23"},"last_update":"2025-06-20T11:27:26+08:00"}]}
	data2 = fastjson.GetString(result.Bytes(), "adcode")
	if "0" == fastjson.GetString(result.Bytes(), "results", "0", "now", "code") {
		data2 = ""
		data2 += fastjson.GetString(result.Bytes(), "results", "0", "location", "name")
		data2 += " 天气："
		data2 += fastjson.GetString(result.Bytes(), "results", "0", "now", "text")
		data2 += " 温度："
		data2 += fastjson.GetString(result.Bytes(), "results", "0", "now", "temperature")
	} else {
		err = errors.New(fastjson.GetString(result.Bytes(), "info"))
	}
	return
}
