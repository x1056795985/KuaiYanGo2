function mt论坛签到(JSON形参文本) {
    局_url = "https://bbs.binmt.cc/plugin.php?id=k_misign:sign&operation=qiandao&formhash=01888762&format=empty&inajax=1&ajaxtarget="

    局_ck = 'cQWy_2132_connect_is_bind=1; cQWy_2132_connect_uin=13243E27EF0E4DF22784290C61761C79; cQWy_2132_smile=5D1; cQWy_2132_client_token=13243E27EF0E4DF22784290C61761C79; cQWy_2132_nofavfid=1; cQWy_2132_saltkey=IO0VlBP7; cQWy_2132_lastvisit=1748389294; cQWy_2132_sid=i47TJW; cQWy_2132_sendmail=1; cQWy_2132_st_p=0%7C1750736296%7Cdfb88c73187eee873591e5692931b4b5; cQWy_2132_visitedfid=40D50D41D42D44D39D38D46D53D2; cQWy_2132_viewid=tid_151861; cQWy_2132_con_request_uri=https%3A%2F%2Fbbs.binmt.cc%2Fconnect.php%3Fmod%3Dlogin%26op%3Dcallback%26referer%3Dforum.php; cQWy_2132_client_created=1750736337; cQWy_2132_ulastactivity=41c4hhIJCUi5JYeq9cAf4fJpWR9eoxNgJ0wJrD2R7CSUAU7%2F9oKP; cQWy_2132_auth=3bbdVgO9xJu82JjBeV2WVHrJpzBvAyD370FNMCECLiTtno7jho8pbQgloAwLeey3GcDo33ejXPyd1wXfiT5YP0RryA; cQWy_2132_connect_login=1; cQWy_2132_stats_qc_login=3; cQWy_2132_lastcheckfeed=73470%7C1750736337; cQWy_2132_onlineusernum=1256; cQWy_2132_lastact=1750736357%09plugin.php%09;'

    局_头信息 = []
    返回对象 = $api_网页访问_GET(局_url, 局_头信息, 局_ck, 15, "")
    return 返回对象.Body //只返回响应信息
}