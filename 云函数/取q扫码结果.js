function 取q扫码结果(JSON形参文本) {
    if (JSON形参文本 == "")(
        JSON形参文本 = '{"appId":"1101817502","ck":"W+Q8NZphU67+oTP8FT8V0elNm5QLO1K1vB284B9/V0Y+uRhDoKIUzqazhO0qWfV82V6GsCYmwpve+1pXXNH/viOa/s/rDlul4GG52vASoAU4ZcYjHfH9/9r2K2B8ffTvEXYFiBJnjDmrD6eNx0WlNtYh54Q5nopusXMOYmdRcctZsgX6FhpodSusAESVs8Cp"}'
    )

    var data = JSON.parse(JSON形参文本);
    var key = "qq1056795985____"
    data.ck = aesDecrypt(data.ck, key) //qrsig=d38a1b14c6a6480514bce9aa236b1bd6ada72d83f81158c87c5f5b8e98f80a205ab55fb4053430299645b5cfd67a34b5a6ee7bee1010feb3a575a49172504281

    let str = getTextBetweenCharacters(data.ck, "=", ";")[0]

    let ptqrtoken = hash33(str)


    let 局_url = "https://xui.ptlogin2.qq.com/ssl/ptqrlogin?u1=http%3A%2F%2Fconnect.qq.com&from_ui=1&type=1&ptlang=2052&ptqrtoken=" + ptqrtoken + "&daid=381&aid=716027609&pt_3rd_aid=" + data.appId + "&pt_openlogin_data=pt_enable_pwd%3D1%26appid%3D716027609%26pt_3rd_aid%3D" + data.appId + "%26daid%3D381%26pt_skey_valid%3D0%26style%3D35%26force_qr%3D1%26autorefresh%3D1%26s_url%3Dhttp%253A%252F%252Fconnect.qq.com%26refer_cgi%3Dm_authorize%26ucheck%3D1%26fall_to_wv%3D1%26status_os%3D7.1.2%26redirect_uri%3Dauth%253A%252F%252Ftauth.qq.com%252F%26client_id%3D" + data.appId + "%26pf%3Dopenmobile_android%26response_type%3Dtoken%26scope%3Dall%26sdkp%3Da%26sdkv%3D3.5.14.lite%26sign%3DD9BF76568C6053853742E942CEC6845E%26status_machine%3DPixel%26switch%3D1%26time%3D1721368532%26loginfrom%3Dmain%26h5sig%3DEKtTRH-rvRSuVoqzIBhhoc3LcEr5MIP5_Z9zUSBqWM4%26loginty%3D6%26pt_flex%3D1%26loginfrom%3Dmain%26h5sig%3DEKtTRH-rvRSuVoqzIBhhoc3LcEr5MIP5_Z9zUSBqWM4%26loginty%3D6%26&device=2&ptopt=1&pt_uistyle=35&jsver=v1.55.0&aegis_uid=3fbe7f000001ebec-ba49df665031829d-296&r=0.2930985401071531"
    头 = [
        'Accept: */*',
        'Accept-Language: zh-cn',
        'Connection: Keep-Alive',
        'Content-Type: text/plain; Charset=UTF-8',
        'Cookie: qrsig=d38a1b14c6a6480514bce9aa236b1bd6ada72d83f81158c87c5f5b8e98f80a205ab55fb4053430299645b5cfd67a34b5a6ee7bee1010feb3a575a49172504281',
        'Host: xui.ptlogin2.qq.com',
        'Referer: https://xui.ptlogin2.qq.com/ssl/ptqrlogin?u1=http%3A%2F%2Fconnect.qq.com&from_ui=1&type=1&ptlang=2052&ptqrtoken=841587122&daid=381&aid=716027609&pt_3rd_aid=1101817502&pt_openlogin_data=pt_enable_pwd%3D1%26appid%3D716027609%26pt_3rd_aid%3D1101817502%26daid%3D381%26pt_skey_valid%3D0%26style%3D35%26force_qr%3D1%26autorefresh%3D1%26s_url%3Dhttp%253A%252F%252Fconnect.qq.com%26refer_cgi%3Dm_authorize%26ucheck%3D1%26fall_to_wv%3D1%26status_os%3D7.1.2%26redirect_uri%3Dauth%253A%252F%252Ftauth.qq.com%252F%26client_id%3D1101817502%26pf%3Dopenmobile_android%26response_type%3Dtoken%26scope%3Dall%26sdkp%3Da%26sdkv%3D3.5.14.lite%26sign%3DD9BF76568C6053853742E942CEC6845E%26status_machine%3DPixel%26switch%3D1%26time%3D1721368532%26loginfrom%3Dmain%26h5sig%3DEKtTRH-rvRSuVoqzIBhhoc3LcEr5MIP5_Z9zUSBqWM4%26loginty%3D6%26pt_flex%3D1%26loginfrom%3Dmain%26h5sig%3DEKtTRH-rvRSuVoqzIBhhoc3LcEr5MIP5_Z9zUSBqWM4%26loginty%3D6%26&device=2&ptopt=1&pt_uistyle=35&jsver=v1.55.0&aegis_uid=3fbe7f000001ebec-ba49df665031829d-296&r=0.2930985401071531', 'User-Agent: Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36'
    ]
    let 返回对象 = $api_网页访问_POST(局_url, "", 头, data.ck, 15, )

    返回对象 = $api_文本_取出中间文本(返回对象.Body.replace(/\s*/g, ''), "('", "')")


    返回对象 = 返回对象.split("','")
    if (返回对象.lenth < 6) {
        return {
            status: 50,
            msg: '访问失败'
        }
    }
    //返回对象 = $api_网页访问_POST(局_url, "api=123", 15, "")
    //{"StatusCode":200,"Headers":"Date: Sun, 21 May 2023 10:26:32 GMT\r\nContent-Length: 0\r\nContent-Type: application/x-www-form-urlencoded,\r\n","Cookies":"","Body":""}
    return {
        status: parseInt(返回对象[0]),
        auth: 返回对象[2] ? aesEncrypt(返回对象[2], key) : "",
        msg: 返回对象[4]
    }
}

function hash33(e) {
    for (var t = 0, n = 0, o = e.length; n < o; ++n) t += (t << 5) + e.charCodeAt(n);
    return 2147483647 & t
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
//截取两个字符中间的字符串
function getTextBetweenCharacters(str, char1, char2) {
    const regex = new RegExp(char1 + "(.*?)" + char2, "g");
    let matches = str.match(regex);
    if (matches) {
        return matches.map((match) =>
            match.replace(char1, "").replace(char2, "")
        );
    }
    return [];
}
import "https://lf3-cdn-tos.bytecdntp.com/cdn/expire-1-M/crypto-js/3.1.9/crypto-js.js" //网络请求加载  会读取本地缓存 缓存目录 ./云函数/lib/网址+路径.....