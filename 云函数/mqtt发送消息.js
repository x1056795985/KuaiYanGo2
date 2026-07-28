function mqtt发送消息(JSON形参文本) {

    let 主题 = "13111111551_1";
    let 消息 = "飞鸟快验";
    var 局_结果对象 = $api_mqtt发送消息(主题, 消息)

    if (局_结果对象.IsOk) {
        return "ok"
    }
    return 局_结果对象.Err //这里存放错误信息,只有参数不正确,或数据库无法连接的情况才会有错误
}