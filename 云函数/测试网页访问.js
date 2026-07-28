function 测试网页访问(JSON形参文本) {


    局_url = "https://211.156.200.95:9017/bgztService"
    局_post = `method=getDSPostInfo&service=InPostService&data={"V_JGBH":"1","V_YJHM":"1111111111","V_SJRDH":"","C_YJZT":"0","V_CZYGH":"372622","C_QRJS":0,"V_ACCOUNT":"","C_DXFSBZ":"0","C_RKMS":"3","V_CNDF":"0","C_SFPPHISCUS":"0","V_BBH":"1.9.9"}`
    协议头 = [
        "Accept: */*",
        "Accept-Language: zh-cn",
        "Connection: Keep-Alive",
        "Content-Type: application/x-www-form-urlencoded",
        "Host: 211.156.200.95:9017",
        "Referer: https://211.156.200.95:9017/bgztService",
        "User-Agent: Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36",
    ]


    返回对象 = $api_网页访问_POST(局_url, 局_post, 协议头.join("\r"), "", 15, "")
    //{"StatusCode":200,"Headers":"Date: Sun, 21 May 2023 10:26:32 GMT\r\nContent-Length: 0\r\nContent-Type: application/x-www-form-urlencoded,\r\n","Cookies":"","Body":""}
    return 返回对象.Body //只返回响应信息
}