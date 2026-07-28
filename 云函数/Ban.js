function Ban(卡号) {


    var result = $api_执行SQL功能("UPDATE db_Ka SET Status = 2 WHERE Name = '" + 卡号 + "'");
    if (result.isOk) {
        return Number(result.Err); // 返回影响行数
    }
    return result.Err;
}