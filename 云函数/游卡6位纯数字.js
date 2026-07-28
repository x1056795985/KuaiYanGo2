function 游卡6位纯数字(JSON形参文本) {
    var 局_软件用户信息 = $api_取软件用户详情($用户在线信息)

    if (局_软件用户信息.VipNumber < 0.00) {
        return {
            Code: -1,
            Msg: "余额不足a",
        }
    }

    let 任务类型ID = 1
    let 结果 = $api_任务池_任务创建($用户在线信息, 任务类型ID, JSON形参文本)

    if (结果.IsOk) {
        let 局_任务对象 = 结果.Data
        let 任务结果
        for (let i = 0; i < 3; i++) {
            $程序_延时(5000); // 等待1秒
            任务结果 = $api_任务池_任务查询(局_任务对象.TaskUuid)
            if (任务结果.Data.Status !== 1 && 任务结果.Data.Status !== 2 ) { //不是刚创建, 也不是处理中,跳出循环
                break
            }
        }
         //{"IsOk":true,"Err":"","Data":{"ReturnData":"","Status":1,"TimeEnd":0,"TimeStart":1695016978}}
        if (任务结果.Data.Status === 3){
           // 如果是成功,直接返回
            return {
                Code: 1,
                Msg: "ok",
                captchaId: "",
                recognition: 3533,
                md5:  "",
            }
        }
    }



    let 局_形参对象 = JSON.parse(JSON形参文本) //使用JSON.parse() 将JSON字符串转为JS对象;

    let 局_结果 = $api_用户Id增减积分($用户在线信息, "-0.01", "测试公共函数扣积分") //{IsOk: true, Err: ""}
    if (局_结果.IsOk !== true) {
        return {
            Code: -1,
            Msg: "余额不足b",
        }
    }


    let 局_url = "http://upload.chaojiying.net/Upload/Processing.php"
    let 局_响应 = $api_网页访问_POST(局_url, '{"user":"13109812593","pass":"x000000","softid":952706,"codetype":4006,"file_base64":"' + 局_形参对象.file_base64 + '"}', 15, "")
    //{"StatusCode":200,"Headers":"Strict-Transport-Security: max-age=15768000\r\nServer: nginx\r\nVary: Accept-Encoding\r\nContent-Type: application/json\r\nAccess-Control-Allow-Origin: *\r\nDate: Sat, 16 Sep 2023 07:38:02 GMT\r\n","Cookies":"","Body":"{\"err_no\":0,\"err_str\":\"OK\",\"pic_id\":\"1226515380787470158\",\"pic_str\":\"838260\",\"md5\":\"b67c1984913d299e09e6e09f0e423f41\"}"}
    let 超级鹰对象 = JSON.parse(局_响应.Body)
    //let 超级鹰对象 = JSON.parse('{"err_no":0,"err_str":"OK","pic_id":"1226514460787470156","pic_str":"838260","md5":"91dbbc6b9a6c3e16ac6e1b414bff14ec"}')
    //{"err_no":0,"err_str":"OK","pic_id":"1226514460787470156","pic_str":"838260","md5":"91dbbc6b9a6c3e16ac6e1b414bff14ec"}
    //{"err_no":-10061,"err_str":"不是有效的图片文件","pic_id":"0","pic_str":"","md5":""}
    if (超级鹰对象.err_str != "OK") {
        return {
            Code: -1,
            Msg: 超级鹰对象.err_str,
        }
    }

    return {
        Code: 1,
        Msg: 超级鹰对象.err_str,
        captchaId: 超级鹰对象.pic_id,
        recognition: 超级鹰对象.pic_str,
        md5: 超级鹰对象.md5,
    }

}

const js对象_通用返回 = {
    IsOk: false,
    Err: "",
    Data: {}
};