function 置用户云配置(JSON形参文本) {
    //return $用户在线信息; // {"Key":"aaaaaa","Status":1,"Tab":"AMD Ryzen 7 6800H with Radeon Graphics         |178BFBFF00A40F41","Uid":21,"User":"aaaaaa"}
    //return $应用信息 // {"AppId":10001,"AppName":"演示对接账密限时Rsa交换密匙","Status":3,"VipData":"{\n\"VipData\":\"这里的数据,只有登录成功并且账号会员不过期才会传输出去的数据\",\n\"VipData2\":\"这里的数据,只有登录成功并且账号会员不过期才会传输出去的数据\"\n}
    //return $用户在线信息.Uid

    let 配置名 = "窗口宽度";
    let 配置值 = "360px";

    // $用户在线信息.LoginAppid=10001    //通过修改登录appid,可以读取其他应用的云配置信息
    // 局uid=$api_用户名或卡号取uid($用户在线信息.LoginAppid,"aaaaaa")  // 可以通过这个获取指定用户的uid, 其实就是应用用户的来源id  无该用户返回0
    $用户在线信息.Uid = 57 //通过修改uid 用户或卡号id,可以读取其他用户的云配置信息


    var 局_结果对象 = $api_置用户云配置($用户在线信息, 配置名, 配置值)
    if (局_结果对象.IsOk) {
        return "写入成功"
    }

    return 局_结果对象.Err //这里存放错误信息,
}