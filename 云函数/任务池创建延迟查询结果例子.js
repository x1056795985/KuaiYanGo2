function 任务池创建查询例子(形参) {
    let 任务类型ID = 1
    let 结果 = $api_任务池_任务创建($用户在线信息, 任务类型ID, JSON形参文本)
    //{"IsOk":true,"Err":"","Data":{"TaskUuid":"1fb701a9-05c5-442a-8bcc-34bda07050ae"}}

    if (结果.IsOk) {
        let 局_任务对象 = 结果.Data
        let 任务结果
        for (let i = 0; i < 3; i++) {
            $程序_延时(5000); // 等待1秒
            任务结果 = $api_任务池_任务查询(局_任务对象.TaskUuid)
            if (任务结果.Data.Status !== 1 && 任务结果.Data.Status !== 2) { //不是刚创建, 也不是处理中,跳出循环
                break
            }
        }
        //{"IsOk":true,"Err":"","Data":{"ReturnData":"","Status":1,"TimeEnd":0,"TimeStart":1695016978}}
        if (任务结果.Data.Status === 3) {
            // 如果是成功,直接返回
            return {
                Code: 1,
                Msg: "ok",
                recognition: 任务结果.Data.ReturnData,
            }
        }
    }
    return {
        Code: -1,
        Msg: "失败",
    }
}

const js对象_通用返回 = { //api函数返回基本都是这个
    IsOk: false,
    Err: "",
    Data: {}
};