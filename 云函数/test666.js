function test666(形参) {


    var 局_用户信息 = $api_用户Id取详情($用户在线信息) //Id":0,"AgentDiscount":0,"LoginAppid":10000,"LoginIp":"","LoginTime":1519454315,"RegisterIp":"113.235.144.55","RegisterTime":1519454315}

    var 局_软件用户信息 = $api_取软件用户详情($用户在线信息)
    $api_置公共变量(局_用户信息["RegisterIp"], 局_用户信息["RegisterIp"])
    if ($api_读公共变量(局_用户信息["RegisterIp"]) == "") {
        $api_置公共变量(局_用户信息["RegisterIp"], 局_用户信息["RegisterIp"])
        return "RegisterIp为空"
    } else {
        局_结果 = $api_用户Id增减时间点数($应用信息.AppId, $用户在线信息, -3600, "IP试用过多")
        if (局_结果.IsOk) {
            return "ok"
        }

        return 局_结果.Err
    }

}