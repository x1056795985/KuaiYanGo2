function 测试归属js(JSON形参文本) {
    //return $用户在线信息; // {"Key":"aaaaaa","Status":1,"Tab":"AMD Ryzen 7 6800H with Radeon Graphics         |178BFBFF00A40F41","Uid":21,"User":"aaaaaa"}
    //return $应用信息 // {"AppId":10001,"AppName":"演示对接账密限时Rsa交换密匙","Status":3,"VipData":"{\n\"VipData\":\"这里的数据,只有登录成功并且账号会员不过期才会传输出去的数据\",\n\"VipData2\":\"这里的数据,只有登录成功并且账号会员不过期才会传输出去的数据\"\n}
    //return $用户在线信息.Uid

    var 局_用户信息 = $api_用户Id取详情($用户在线信息) //{"Id":21,"User":"aaaaaa","PassWord":"af15d5fdacd5fdfea300e88a8e253e82","Phone":"13109812593","Email":"1056795985@qq.com","Qq":"1059795985","SuperPassWord":"af15d5fdacd5fdfea300e88a8e253e82","Status":1,"Rmb":91.39,"Note":"","RealNameAttestation":"","Role":0,"UPAgentId":0,"AgentDiscount":0,"LoginAppid":10000,"LoginIp":"","LoginTime":1519454315,"RegisterIp":"113.235.144.55","RegisterTime":1519454315}
    return 局_用户信息
}