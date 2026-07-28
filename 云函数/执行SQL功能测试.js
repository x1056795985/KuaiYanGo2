function 执行SQL功能测试(JSON形参文本) {


    var 局_结果对象 = $api_执行SQL功能("UPDATE db_public_js SET Note=CONCAT(Note, ?) WHERE  Id=11") //获取公共函数数据库全部信息

    //带预处理绑定参数方式
    var data = [
        "追加文本",
        11
    ]
    var 局_结果对象 = $api_执行SQL功能("UPDATE db_public_js SET Note=CONCAT(Note, ?) WHERE  Id = ?", data) //获取公共函数数据库全部信息


    if (局_结果对象.IsOk) {
        //这里说明成功了,
        let 影响行数 = Number(局_结果对象.Err)
        return 影响行数 //返回影响行数
    }

    return 局_结果对象.Err

}