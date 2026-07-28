function Hook_ApiHOOk执行前例子888(JSON请求明文) {
    //这里的错误无法拦截,所以,如果js错误,可能会导致,用户返回"Api不存在"
    //JSON.stringify($Request)  //在 $Request里可以获取到 请求的大部分信息
    //{"Method":"POST","Url":{"Scheme":"","Opaque":"","User":null,"Host":"","Path":"/Api","RawPath":"","OmitHost":false,"ForceQuery":false,"RawQuery":"AppId=10002","Fragment":"","RawFragment":""},"Header":["Connection: Keep-Alive","Referer: http://127.0.0.1:18888/Api?AppId=10002","Content-Length: 467","Content-Type: application/x-www-form-urlencoded; Charset=UTF-8","Accept: */*","Accept-Language: zh-cn","User-Agent: Mozilla/4.0 (compatible; MSIE 9.0; Windows NT 6.1)","Token: PNYDKXDHLORTNVGEEY99YYSPQGFLQF7L"],"Host":"127.0.0.1:18888","Body":[]}

    //局_url = "https://www.baidu.com/"
    //局_返回 = $api_网页访问_GET(局_url, 15, "")
    //局_返回 = $api_网页访问_POST(局_url, "api=123", 15, "")
    //{"StatusCode":200,"Headers":"Date: Sun, 21 May 2023 10:26:32 GMT\r\nContent-Length: 0\r\nContent-Type: application/x-www-form-urlencoded,\r\n","Cookies":"","Body":""}

    if (局_返回.Body !== "") {
        //$拦截原因 = "百度可以访问,所以不能登录." 
    }
	//这里可以替换请求明文信息,可以实现很多功能,比如自写算法解密
    return JSON请求明文
}