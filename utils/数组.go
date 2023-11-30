package utils

import "strings"

func S数组_整数是否存在(数组 []int, 整数 int) bool {
	for _, num := range 数组 {
		if num == 整数 {
			return true
		}
	}
	return false
}

// 判断数组各元素是否是空字符串或空格
func S数组_是否为空(list []string) (isEmpty bool) {

	if len(list) == 0 {
		return true
	}

	isEmpty = true
	for _, f := range list {

		if strings.TrimSpace(f) != "" {
			isEmpty = false
			break
		}
	}

	return isEmpty
}
