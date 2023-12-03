package utils

import (
	"sort"
	"strings"
)

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

func S数组_排序整数(arr []int) []int {
	局_arr := arr
	sort.Ints(局_arr)
	return 局_arr
}

func S数组_排序文本(arr []string) []string {
	局_arr := arr
	sort.Strings(局_arr)
	return 局_arr
}
