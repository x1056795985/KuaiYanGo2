function Hook_Api例子(JSON请求明文) {
    //JSON.stringify($Request)  //在 $Request里可以获取到 请求的大部分信息
    //{"Method":"POST","Url":{"Scheme":"","Opaque":"","User":null,"Host":"","Path":"/Api","RawPath":"","OmitHost":false,"ForceQuery":false,"RawQuery":"AppId=10002","Fragment":"","RawFragment":""},"Header":["Connection: Keep-Alive","Referer: http://127.0.0.1:18888/Api?AppId=10002","Content-Length: 467","Content-Type: application/x-www-form-urlencoded; Charset=UTF-8","Accept: */*","Accept-Language: zh-cn","User-Agent: Mozilla/4.0 (compatible; MSIE 9.0; Windows NT 6.1)","Token: PNYDKXDHLORTNVGEEY99YYSPQGFLQF7L"],"Host":"127.0.0.1:18888","Body":[]}

    局_url = "https://www.baidu.com/"
    局_返回 = $api_网页访问_GET(局_url, 15, "")
    //局_返回 = $api_网页访问_POST(局_url, "api=123", 15, "")
    //{"StatusCode":200,"Headers":"Date: Sun, 21 May 2023 10:26:32 GMT\r\nContent-Length: 0\r\nContent-Type: application/x-www-form-urlencoded,\r\n","Cookies":"","Body":""}

    if (局_返回.Body !== "") {
        $拦截原因 = "百度可以访问,所以不能登录." + JSON.stringify($Request)
    }

    return JSON明文
}