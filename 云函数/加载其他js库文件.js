function 加载其他js库文件(JSON形参文本) {

    return CryptoJS.MD5("ok").toString();
}

//import "https://lf9-cdn-tos.bytecdntp.com/cdn/expire-1-M/crypto-js/4.1.1/crypto-js.min.js" //网络请求加载  会读取本地缓存 缓存目录 ./云函数/lib/网址+路径.....
//@import "https://lf9-cdn-tos.bytecdntp.com/cdn/expire-1-M/crypto-js/4.1.1/crypto-js.min.js" //网络请求加载 不会读取本地缓存,影响速度不推荐,测试可以,
import "lib/lf9-cdn-tos.bytecdntp.com/cdn/expire-1-M/crypto-js/4.1.1/crypto-js.min.js" //加载本地js文件 实际读取路径 ./云函数/crypto-js.min.js    所有要加载的本地js文件都要放到  运行目录/云函数/  内,可以加文件夹