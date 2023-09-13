package utils

import (
	"time"
)

func S时间_延迟(秒数 int) bool {
	for i := 0; i < 秒数; {
		time.Sleep(1 * time.Second)
	}
	return true
}

// S时间_文本到时间戳
// @时间文本  2006-01-02 15:04:05
func S时间_文本到时间戳(时间文本 string) int {

	formatTime, err := time.ParseInLocation("2006-01-02 15:04:05", 时间文本, time.Local) //使用parseInLocation将字符串格式化返回本地时区时间
	if err == nil {
		return int(formatTime.Unix())
	}
	return 0

}

// S时间_取现行时间戳13
func S时间_取现行时间戳13() int64 {
	return time.Now().UnixNano() / 1e6
}
