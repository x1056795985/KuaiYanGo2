function uuid添加到队列(JSON形参文本) {
    var data = JSON.parse(JSON形参文本);
    if (!data.uuid)(
        data = {
            "uuid": "33dde03f-37a4-4454-bc04-7e273aeb5ab8"
        }
    )
    局_结果 = $api_任务池Uuid添加到队列(data.uuid)
    return 局_结果
}