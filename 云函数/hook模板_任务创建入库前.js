function hook模板_任务创建入库前(任务JSON格式参数) {
    /*
    return $用户在线信息; // {"Key":"aaaaaa","Status":1,"Tab":"AMD Ryzen 7 6800H with Radeon Graphics         |178BFBFF00A40F41","Uid":21,"User":"aaaaaa"}
    return $应用信息 // {"AppId":10001,"AppName":"演示对接账密限时Rsa交换密匙","Status":3,"VipData":"..."}
    return $用户在线信息.Uid
    var 局_用户信息 = $api_用户Id取详情($用户在线信息)
    例子随机 拦截任务提交
    任务JSON格式参数 = 任务JSON格式参数.replace(/'/g, '"')
    var 局_形参对象 = JSON.parse(任务JSON格式参数);
    局_结果 = $api_用户Id增减余额($用户在线信息, -局_形参对象.a, "测试任务池Hook内扣余额")
    if (!局_结果.IsOk) {
        $拦截原因 = "扣费失败" + 局_结果.Err
    }
    if (Math.floor(Math.random() * 10) > 5) {
        $拦截原因 = "如果值不为空,则任务拦截,响应拦截原因"
    }
    */
    return 任务JSON格式参数
}