function 测试1111(JSON形参文本) {

    JSON形参文本 = "eyLlhbPplK7or40iOiIlRTYlOTklQjQlRTUlQTQlQTkiLCJuIjoxfQ=="

    return $api_编码_BASE64解码(JSON形参文本)
    //return $用户在线信息; // {"Key":"aaaaaa","Status":1,"Tab":"AMD Ryzen 7 6800H with Radeon Graphics         |178BFBFF00A40F41","Uid":21,"User":"aaaaaa"}
    //return $应用信息 // {"AppId":10001,"AppName":"演示对接账密限时Rsa交换密匙","Status":3,"VipData":"{\n\"VipData\":\"这里的数据,只有登录成功并且账号会员不过期才会传输出去的数据\",\n\"VipData2\":\"这里的数据,只有登录成功并且账号会员不过期才会传输出去的数据\"\n}

    // $用户在线信息.Id = 1809267
    // var 局_结果对象 = $api_在线注销($用户在线信息)
    // if (局_结果对象.IsOk) {
    //     return 局_结果对象
    // }
    // return 局_结果对象
    $用户在线信息.User = "aaaaaa"
    $用户在线信息.LoginAppid = 10001
    $用户在线信息.Uid = 0

    局_结果 = $api_用户Id增减时间点数(10001, $用户在线信息, -3600, "测试扣时间")
    //局_结果 = $api_用户Id增减积分($用户在线信息, 0.01, "测试公共函数扣余额")
    return 局_结果 // {IsOk: true, Err: ""}
    var 局_用户信息 = $api_用户Id取详情($用户在线信息) //{"Id":21,"User":"aaaaaa","PassWord":"af15d5fdacd5fdfea300e88a8e253e82","Phone":"13109812593","Email":"1056795985@qq.com","Qq":"1059795985","SuperPassWord":"af15d5fdacd5fdfea300e88a8e253e82","Status":1,"Rmb":91.39,"Note":"","RealNameAttestation":"","Role":0,"UPAgentId":0,"AgentDiscount":0,"LoginAppid":10000,"LoginIp":"","LoginTime":1519454315,"RegisterIp":"113.235.144.55","RegisterTime":1519454315}
    return 局_用户信息
}