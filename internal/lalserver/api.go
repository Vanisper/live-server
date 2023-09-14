package lalserver

import (
	"fmt"
	"live-server/library/LAL/pkg/base"
	"live-server/pkg/utils/encoding"

	"github.com/imroc/req/v3"
)

var (
	httpApiEnable     = true
	httpApiPort       = "8083"
	httpApiRequestUrl = fmt.Sprintf("http://%s:%s", baseUrl, httpApiPort)
	// /api/stat/group
	httpApiStatGroupUrl = fmt.Sprintf("%s/api/stat/group", httpApiRequestUrl)
	// /api/stat/all_group
	httpApiStatAllGroupUrl = fmt.Sprintf("%s/api/stat/all_group", httpApiRequestUrl)

	// /api/ctrl/start_relay_pull
	httpApiCtrlStartRelayPullUrl = fmt.Sprintf("%s/api/ctrl/start_relay_pull", httpApiRequestUrl)
	// /api/ctrl/stop_relay_pull
	httpApiCtrlStopRelayPullUrl = fmt.Sprintf("%s/api/ctrl/stop_relay_pull", httpApiRequestUrl)
)

/*
接口规则

1. 所有接口的返回结果中，必含的一级参数：

	{
		"error_code": 0,
		"desp": "succ",
		"data": ...
	}

2. error_code列表：

error_code	desp	说明
0	succ	调用成功
1001	group not found	group不存在
1002	param missing	必填参数缺失
1003	session not found	session不存在
2001	多种值，表示失败的具体原因	start_relay_pull失败
2002	打开gb28181端口失败	start_rtp_pub
*/

// /api/stat/group: curl http://127.0.0.1:8083/api/stat/group?stream_name=test110
func HttpApiStatGroup(stream_name string) base.ApiStatGroupResp {
	res := req.C().R().SetQueryParams(map[string]string{
		"stream_name": stream_name,
	}).MustGet(httpApiStatGroupUrl)

	var ret base.ApiStatGroupResp
	encoding.JSONDecode(res.String(), &ret)
	return ret
}

// /api/stat/all_group: curl http://127.0.0.1:8083/api/stat/all_group
func HttpApiStatAllGroup() base.ApiStatAllGroupResp {
	res := req.C().R().MustGet(httpApiStatAllGroupUrl)

	var ret base.ApiStatAllGroupResp
	encoding.JSONDecode(res.String(), &ret)
	return ret
}

// /api/ctrl/start_relay_pull: curl -H "Content-Type:application/json" -X POST -d '{"url": "rtmp://127.0.0.1/live/test110?token=aaa&p2=bbb", "pull_retry_num": 0}' http://127.0.0.1:8083/api/ctrl/start_relay_pull
func HttpApiCtrlStartRelayPull(url string, stream_name string) base.ApiCtrlStartRelayPullResp {
	res := req.C().R().SetBody(map[string]interface{}{
		"url":         url,
		"stream_name": stream_name,
	}).MustPost(httpApiCtrlStartRelayPullUrl)

	var ret base.ApiCtrlStartRelayPullResp
	encoding.JSONDecode(res.String(), &ret)
	return ret
}

// /api/ctrl/stop_relay_pull: curl http://127.0.0.1:8083/api/ctrl/stop_relay_pull?stream_name=test110
func HttpApiCtrlStopRelayPull(stream_name string) base.ApiCtrlStopRelayPullResp {
	res := req.C().R().SetQueryParams(map[string]string{
		"stream_name": stream_name,
	}).MustGet(httpApiCtrlStopRelayPullUrl)
	var ret base.ApiCtrlStopRelayPullResp
	encoding.JSONDecode(res.String(), &ret)
	return ret
}
