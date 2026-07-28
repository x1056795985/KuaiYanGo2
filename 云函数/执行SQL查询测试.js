function 执行SQL查询测试(JSON形参文本) {

    //直接执行sql 
    var 局_结果对象 = $api_执行SQL查询(" SELECT * FROM db_public_js")

    //带预处理绑定参数
    var data = [
        [5, 6],
        1
    ]
    //var 局_结果对象 = $api_执行SQL查询(" SELECT * FROM db_public_js WHERE id IN ? and IsVip = ?", data)
    var 局_结果对象 = $api_执行SQL查询(" SELECT * FROM db_public_js")
    if (局_结果对象.IsOk) {
        //这里说明查询成功了,
        return 局_结果对象.Err
    }
    //return 局_结果对象.Err   //这个会把结果返回的文本
    return 局_结果对象.Data //这个会把结果转换成对象

}