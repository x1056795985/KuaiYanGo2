function 任务池队列监控(JSON形参文本) {
    let ret = $api_任务池_取队列长度()

    // {"Return":{"IsOk":true,"Err":"成功","Data":{"9":"8","10":"16"}},"Time":8860}
    if (ret.IsOk) {
        let data = ret.Data
        if (data["9"] > 50) {
            return wx通知("tid5队列异常" + String(data["9"]), "tid5队列异常" + String(data["9"]))
        }
    }

    return ret
}

function wx通知(标题, 内容) {
    data = {
        "token": "2729b76d43f1477f908603a9b6bf0321",
        "title": 标题,
        "content": 内容,
        "template": "html",
        "channel": "wechat",
        "pre": ""
    }
    return $api_网页访问_POST("https://www.pushplus.plus/send", JSON.stringify(data), "", "", 15, "").Body
}