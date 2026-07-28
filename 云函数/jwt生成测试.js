function jwt生成测试(JSON形参文本) {
    $应用信息.AppId = 10001 //测试数据
    $用户在线信息.Uid = 2 //测试数据

    var 局_AppUser = $api_用户Id取详情($用户在线信息)
    let jwtMap = JSON.parse(JSON.stringify(局_AppUser)) //因为局_AppUser实际是结构体,不是js对象, 所以无法赋值iat,必须转换成js的对象,才能添加新成员
    jwtMap.iat = Math.floor(Date.now() / 1000) //添加一个生成时间
    let 签名密钥 = "1234567890abcdefg" //这个是签名密钥,不能存放在客户端,必须是可信服务器,才可以存储,用来校验jwt
    var 局_用户信息 = $api_Jwt生成(JSON.stringify(jwtMap), 签名密钥)
    return 局_用户信息
}