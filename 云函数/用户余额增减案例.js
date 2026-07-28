function 用户余额增减案例(JSON形参文本) {
    JSON形参文本 = JSON形参文本.replace(/'/g, '"') //因为易语言 双引号不方便,所以到js里换成替换单引号成双引号 //注意永远不要相信客户端传参

    var 局_形参对象 = JSON.parse(JSON形参文本); //使用JSON.parse() 将JSON字符串转为JS对象;

    if (局_形参对象.a > 0) {
        $拦截原因 = "金额不能大于0"
        return {
            IsOk: false,
            Err: "金额不能大于0"
        }
    } else {
        局_结果 = $api_用户Id增减余额($用户在线信息, 局_形参对象.a, "测试公共函数扣余额")
    }


    return 局_结果 // {IsOk: true, Err: ""}
}