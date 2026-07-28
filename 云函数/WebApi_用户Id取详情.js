function WebApi_用户Id取详情(JSON形参文本) {
    JSON形参文本 = JSON形参文本.replace(/'/g, '"') //因为易语言 双引号不方便,所以到js里换成替换单引号成双引号 //注意永远不要相信客户端传参

    var 局_形参对象 = JSON.parse(JSON形参文本); //使用JSON.parse() 将JSON字符串转为JS对象;

    $用户在线信息.Uid = 0 //下边这个传对象,所以先赋值Uid 到对象内
    $用户在线信息.User = "aaaaaa"
    $用户在线信息.LoginAppid = 10001
    局_结果 = $api_取软件用户详情($用户在线信息)

    return 局_结果
}