package cache

import (
	"time"
)

// 定义一个缓存接口 缓存实现方式都用这个接口
type Cache interface {
	Set(k string, v interface{}, d time.Duration)       // Set 添加cache 无论是否存在都会覆盖
	Get(k string) (interface{}, bool)                   // Get 根据key获取 cache
	Delete(key string)                                  // Delete 删除k的cache 如果 capture != nil 会调用 capture 函数 将 kv传入
	Increment(k string, n int64) error                  // Increment 为k对应的value增加n n必须为数字类型
	Add(k string, x interface{}, d time.Duration) error // Add 添加cache 如果存在则抛出异常 原子性,可以当做锁,添加成功的枪到锁
}
