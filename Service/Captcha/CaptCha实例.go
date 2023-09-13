package Captcha

// 当开启多服务器部署时，替换下面的配置，使用redis共享存储验证码
// var store = captcha.NewDefaultRedisStore()

// var store = base64Captcha.DefaultMemStore
var H缓存验证码校验实例 = H缓存验证码{}
