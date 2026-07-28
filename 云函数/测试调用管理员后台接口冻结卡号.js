function 测试调用管理员后台接口冻结卡号(参数) {
    //详细说明 官网常见问题  http://www.fnkuaiyan.cn/%E6%8C%87%E5%8D%97/%E5%B8%B8%E8%A7%81%E9%97%AE%E9%A2%98.html#%E5%85%AC%E5%85%B1%E5%87%BD%E6%95%B0%E5%86%85token%E8%B0%83%E7%94%A8%E5%90%8E%E5%8F%B0%E6%88%96%E4%BB%A3%E7%90%86%E6%8E%A5%E5%8F%A3%E5%8A%9F%E8%83%BD
    局_url = "http://127.0.0.1:18888/Admin/AppUser/SetStatus"
    局_post = '{"AppId":10001,"Id":[69],"Status":2}' //这里可以根据需求自己修改参数, 这个id是卡号id,AppId是卡号归属id
    局_token = "WD3NMTTWNG40DERXA6WRZTK3BZZLTKMJ"  //这个需要自己抓包替换
    协议头 = "Token: " + 局_token
    返回对象 = $api_网页访问_POST(局_url, 局_post, 协议头, 15, "")

    return 返回对象
}