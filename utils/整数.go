package utils

import "sync/atomic"

// 整数原子递增1 线程安全 返回新值
func Z整数_原子递增(整数变量 *int64) int64 {
	return atomic.AddInt64(整数变量, 1)
}

// 整数原子递增1 线程安全 返回新值
func Z整数_原子递减(整数变量 *int64) int64 {
	return atomic.AddInt64(整数变量, -1)
}
