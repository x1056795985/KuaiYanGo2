function 取q登陆二维码(JSON形参文本) {
    var data = JSON.parse(JSON形参文本);
    if (!data.appId)(
        data = {
            "appId": 1101817502
        }
    )

    let 代理ip = $api_网页访问_GET("http://v2.api.juliangip.com/company/postpay/getips?num=1&pt=2&result_type=text&split=1&trade_no=6096034117538547&sign=5d9b9269e87a835e1ac6a0d8700b091b").Body
    代理ip = 正则_代理ip(代理ip)
    if (代理ip) {
        代理ip = ""
    }

    let timestamp = Date.now();
    let shortTimestamp = timestamp.toString().slice(0, 10); // 从10位毫秒数开始截取，直到末尾
    let 局_url = "https://xui.ptlogin2.qq.com/cgi-bin/xlogin?pt_enable_pwd=1&appid=716027609&pt_3rd_aid=" +
        data.appId + "&daid=381&pt_skey_valid=0&style=35&force_qr=1&autorefresh=1&s_url=http%3A%2F%2Fconnect.qq.com&refer_cgi=m_authorize&ucheck=1&fall_to_wv=1&status_os=11&redirect_uri=auth%3A%2F%2Ftauth.qq.com%2F&client_id=1105412664&pf=openmobile_android&response_type=token&scope=all&sdkp=a&sdkv=3.5.7.lite&sign=f1e1bf720d40eebf1e95d70310bfabf2&status_machine=Redmi+K30+5G+Speed&switch=1&time=1722679468945&show_download_ui=true&h5sig=363wXmww3ftebM39EAw9mOP44xM-8kLXcAB0aJqsnFk&loginty=6"
    let urltest = "https://searchplugin.csdn.net/api/v1/ip/get?ip="

    返回对象 = $api_网页访问_GET(局_url, 取协议头(), "", 15, 代理ip)
    返回对象2 = $api_网页访问_GET(urltest, 取协议头(), "", 15, 代理ip)
    return 代理ip + "|" + 返回对象2.Body + "|" + 返回对象.Body

    let ck = 返回对象.Cookies

    局_url = "https://xui.ptlogin2.qq.com/ssl/ptqrshow?s=8&e=0&appid=716027609&type=0&t=" + shortTimestamp + "&u1=http%3A%2F%2Fconnect.qq.com&daid=381&pt_3rd_aid=" +
        data.appId;
    //局_url = "https://xui.ptlogin2.qq.com/ssl/ptqrshow?s=8&e=0&appid=716027609&type=0&t=0.08790631564002305&u1=http%3A%2F%2Fconnect.qq.com&daid=381&pt_3rd_aid=1105412664"
    返回对象 = $api_网页访问_GET(局_url, 取协议头(), ck, 15, 代理ip)
    let 配置名 = CryptoJS.MD5(返回对象.Cookies).toString()
    let 配置值 = 代理ip;
    let 有效期 = 180; //有效期60秒,超过了自动删除 -1 为永久有效
    var 局_结果对象 = $api_置缓存(配置名, 配置值, 有效期)



    var key = "qq1056795985____"
    var encrypted = aesEncrypt(返回对象.Cookies, key)


    //返回对象 = $api_网页访问_POST(局_url, "api=123", 15, "")
    //{"StatusCode":200,"Headers":"Date: Sun, 21 May 2023 10:26:32 GMT\r\nContent-Length: 0\r\nContent-Type: application/x-www-form-urlencoded,\r\n","Cookies":"","Body":""}
    return {
        "appId": data.appId,
        "ck": encrypted.toString(),
        "qr": 返回对象.Base64Body
    } //只返回响应信息
}

function 取协议头() {
    let aa = [
        "User-Agent: Mozilla/4.0 (compatible; MSIE 9.0; Windows NT 6.1)",
        "Connection: keep-alive",
        "Referer:https://xui.ptlogin2.qq.com/cgi-bin/xlogin?pt_enable_pwd=1&appid=716027609&pt_3rd_aid=1101817502&daid=381&pt_skey_valid=0&style=35&force_qr=1&autorefresh=1&s_url=http%3A%2F%2Fconnect.qq.com&refer_cgi=m_authorize&ucheck=1&fall_to_wv=1&status_os=11&redirect_uri=auth%3A%2F%2Ftauth.qq.com%2F&client_id=1105412664&pf=openmobile_android&response_type=token&scope=all&sdkp=a&sdkv=3.5.7.lite&sign=f1e1bf720d40eebf1e95d70310bfabf2&status_machine=Redmi+K30+5G+Speed&switch=1&time=1722679468945&show_download_ui=true&h5sig=363wXmww3ftebM39EAw9mOP44xM-8kLXcAB0aJqsnFk&loginty=6"
    ]
    return aa.join("\r")

}

function 正则_代理ip(str) {
    // 正则表达式定义，匹配标准的IPv4地址和端口号
    let regex = new RegExp("\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}:\\d+");
    let matchResult = regex.exec(str);
    // 测试字符串中是否存在匹配项
    if (matchResult && matchResult.length > 0) {
        // 返回第一个匹配到的 IP:端口 字符串
        return matchResult[0];
    } else {
        // 如果没有找到匹配项，返回 null 或者提示信息
        return ""; // 或者 return ;
    }
}



//加密
function aesEncrypt(data, mm) {
    var mm = CryptoJS['enc']['Utf8']['parse'](mm);
    var data = CryptoJS['enc']['Utf8']['parse'](data);
    var dataa = CryptoJS['AES']['encrypt'](data, mm, {
        'mode': CryptoJS['mode']['ECB'],
        'padding': CryptoJS['pad']['Pkcs7']
    });
    return dataa['toString']();
}

//解密
function aesDecrypt(data, mm) {
    var mm = CryptoJS['enc']['Utf8']['parse'](mm);
    var dataa = CryptoJS['AES']['decrypt'](data, mm, {
        'mode': CryptoJS['mode']['ECB'],
        'padding': CryptoJS['pad']['Pkcs7']
    });
    return CryptoJS['enc']['Utf8']['stringify'](dataa)['toString']();
}
import "https://lf3-cdn-tos.bytecdntp.com/cdn/expire-1-M/crypto-js/3.1.9/crypto-js.js" //网络请求加载  会读取本地缓存 缓存目录 ./云函数/lib/网址+路径.....