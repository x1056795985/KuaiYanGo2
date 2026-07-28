function vmp计算授权码(JSON形参文本) {
    let Rsa位数 = 1024
    let RsaBase64私钥 = "hjJi7NNcxxZ3Me3syJiRamoCK0kuXtun4JMctTTavf895giOXzMXKaRcW3MxtKGPhT1bAuY8wgMRrfgeNrS6/eJBiK3n06jWy2g04Kcnp5Q/rptrd6YG9j+vcBmScNgaseUQmkT5gc6ujlLo0P3W4oWkf4mJZNJKO0t6+6QEHEE="
    let RsaBase64模数 = "5HSlarnEogn6B/KIOCkVPVPx9s45M9KWFs1lePOY/8szCG2sHe8jalkihKlyQ3b15BAlxeAJ1+0+zTau3tSAUGTJs6s+f2AtLOYWFoyBv1PnNMh2tyDOdIrLYz11VjgN3igD1r6fMA6kpbm2SwUt5EL9MQDpkqfbBSdcBF0NWzE="
    //let base64产品代码 = "js1KvV+j5ys="
    let base64产品代码 = "AAAAAQAAJxE="


    //实测只需要授权一天即可,因为授权码使用后,所有功能不在受时间限制 实际还是需要靠心跳控制时分秒 精准度
    //激活码的到期时间只有激活的时候才检测,被保护的函数执行时不检测,所以登陆后立刻调用,当天有效即可
    //但是为了防止遇到极端11:59:59时间登陆的情况,所以有效时间设置为明天
    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);


    let 授权配置 = {
        UserName: "abc",
        Email: "abc@qq.com",
        ExpireDate: {
            Year: tomorrow.getFullYear(), // 年（如 2024）
            Month: tomorrow.getMonth() + 1, // 月（1-12，需 +1 对齐实际月份）
            Day: tomorrow.getDate() // 日（1-31）
        },
        MaxBuildDate: { //MaxBuildDate 应该是 注册码生成时间,
            Year: new Date().getFullYear(), // 年（如 2024）
            Month: new Date().getMonth() + 1, // 月（1-12，需 +1 对齐实际月份）
            Day: new Date().getDate() // 日（1-31）
        },
        TimeLimit: 1,
        Hwid: "wO1Xs+VL+7afzVpicvaT/YrMrwdO/aBLusH8eg==",
        UserDataBas64: "",
    }


    var 局_结果 = $api_VMP计算授权码(Rsa位数, RsaBase64私钥, RsaBase64模数, base64产品代码, JSON.stringify(授权配置))
    //{"IsOk":true,"Err":"","Data":"WTQVoqFIvquq3wK8PCiAh6B...省略中间的授权码...HG9NIvE/4swJGYrEdRC5ay+SX7Viil2cw="}

    return 局_结果
}