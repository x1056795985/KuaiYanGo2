function 获取用户相关信息(形参) {
    //return $用户在线信息; // {"Key":"aaaaaa","Status":1,"Tab":"AMD Ryzen 7 6800H with Radeon Graphics         |178BFBFF00A40F41","Uid":21,"User":"aaaaaa"}
    //return $应用信息 // {"AppId":10001,"AppName":"演示对接账密限时Rsa交换密匙","Status":3,"VipData":"{\n\"VipData\":\"这里的数据,只有登录成功并且账号会员不过期才会传输出去的数据\",\n\"VipData2\":\"这里的数据,只有登录成功并且账号会员不过期才会传输出去的数据\"\n}
    //return $用户在线信息.Uid

    //var 局_用户信息 = $api_用户Id取详情($用户在线信息) //Id":0,"AgentDiscount":0,"LoginAppid":10000,"LoginIp":"","LoginTime":1519454315,"RegisterIp":"113.235.144.55","RegisterTime":1519454315}
    //var 局_卡号信息 = $api_卡号Id取详情($用户在线信息)

    $用户在线信息.LoginAppid = 10001
    $用户在线信息.User = "aaaaaa"
    $用户在线信息.Uid = 0
    var 局_软件用户信息 = $api_取软件用户详情($用户在线信息)

    //$api_置动态标记($用户在线信息, $用户在线信息.Tab + "追加文本")

    return 局_软件用户信息
}