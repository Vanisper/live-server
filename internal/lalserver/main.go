package lalserver

import (
	"embed"
	"fmt"
	"strconv"
	"strings"

	"live-server/library/LAL/pkg/logic"
	"live-server/library/naza/pkg/nazalog"
	"live-server/pkg/configs"
	"live-server/pkg/utils"
)

//go:embed all:configs
var conf embed.FS
var configPath = "configs"
var cachePath = "caches"

var (
	baseUrl = "127.0.0.1"
	// http(s)://
	hlsHttpPort    = "8080"
	hlsHttpEnable  = true
	hlsHttpsPort   = "4433"
	hlsHttpsEnable = true
	hlsUrlPattern  = "/hls/"
	hlsDefaultUrl  = fmt.Sprintf("http://%s:%s%s", baseUrl, hlsHttpPort, hlsUrlPattern)
	// rtsp(s)://
	rtspPort       = "5544"
	rtspEnable     = true
	rtspsPort      = "5322"
	rtspsEnable    = true
	rtspUrlPattern = "/live/"
	rtspDefaultUrl = fmt.Sprintf("rtsp://%s:%s%s", baseUrl, rtspPort, rtspUrlPattern)
)

func HlsDefaultUrl(name string) string {
	return fmt.Sprintf("%s%s.m3u8", hlsDefaultUrl, name)
}

func RtspDefaultUrl(name string) string {
	return fmt.Sprintf("%s%s", rtspDefaultUrl, name)
}

func Init() {
	// 初始化配置
	configs.InitConfig(configPath, &conf)
	afterConf()

	// 启动lalserver
	defer nazalog.Sync()
	confFilename := "configs/lalserver.conf.json"
	lals := logic.NewLalServer(func(option *logic.Option) {
		option.ConfFilename = confFilename
	})
	err := lals.RunLoop()
	nazalog.Infof("lal server loop done. err=%+v", err)
}

// afterConf 配置文件加载后的操作
func afterConf() {
	newStr := strings.Replace(confRawContent, "{{confPath}}", configPath, -1)
	newStr = strings.Replace(newStr, "{{cachePath}}", cachePath, -1)
	newStr = strings.Replace(newStr, "{{baseUrl}}", baseUrl, -1)
	// http api 用于控制或查询视频流
	newStr = strings.Replace(newStr, "{{httpApiEnable}}", strconv.FormatBool(httpApiEnable), -1)
	newStr = strings.Replace(newStr, "{{httpApiPort}}", httpApiPort, -1)
	// hsl  http(s)://*.m3u8
	newStr = strings.Replace(newStr, "{{hlsHttpPort}}", hlsHttpPort, -1)
	newStr = strings.Replace(newStr, "{{hlsHttpEnable}}", strconv.FormatBool(hlsHttpEnable), -1)
	newStr = strings.Replace(newStr, "{{hlsHttpsPort}}", hlsHttpsPort, -1)
	newStr = strings.Replace(newStr, "{{hlsHttpsEnable}}", strconv.FormatBool(hlsHttpsEnable), -1)
	newStr = strings.Replace(newStr, "{{hlsUrlPattern}}", hlsUrlPattern, -1)
	// rtsp  rtsp(s)://*
	newStr = strings.Replace(newStr, "{{rtspPort}}", rtspPort, -1)
	newStr = strings.Replace(newStr, "{{rtspEnable}}", strconv.FormatBool(rtspEnable), -1)
	newStr = strings.Replace(newStr, "{{rtspsPort}}", rtspsPort, -1)
	newStr = strings.Replace(newStr, "{{rtspsEnable}}", strconv.FormatBool(rtspsEnable), -1)

	err := utils.WriteToFile(newStr, configPath+"/lalserver.conf.json", true)
	if err != nil {
		panic(err)
	}
}

var confRawContent = `{
	"# doc of config": "https://pengrl.com/lal/#/ConfigBrief",
	"conf_version": "v0.4.1",
	"rtmp": {
	  "enable": true,
	  "addr": ":1935",
	  "rtmps_enable": true,
	  "rtmps_addr": ":4935",
	  "rtmps_cert_file": "./{{confPath}}/cert.pem",
	  "rtmps_key_file": "./{{confPath}}/key.pem",
	  "gop_num": 0,
	  "single_gop_max_frame_num": 0,
	  "merge_write_size": 0
	},
	"in_session": {
	  "add_dummy_audio_enable": false,
	  "add_dummy_audio_wait_audio_ms": 150
	},
	"default_http": {
	  "http_listen_addr": ":{{hlsHttpPort}}",
	  "https_listen_addr": ":{{hlsHttpsPort}}",
	  "https_cert_file": "./{{confPath}}/cert.pem",
	  "https_key_file": "./{{confPath}}/key.pem"
	},
	"httpflv": {
	  "enable": true,
	  "enable_https": true,
	  "url_pattern": "/",
	  "gop_num": 0,
	  "single_gop_max_frame_num": 0
	},
	"hls": {
	  "enable": {{hlsHttpEnable}},
	  "enable_https": {{hlsHttpsEnable}},
	  "url_pattern": "{{hlsUrlPattern}}",
	  "out_path": "./{{cachePath}}/hls/",
	  "fragment_duration_ms": 3000,
	  "fragment_num": 6,
	  "delete_threshold": 6,
	  "cleanup_mode": 2,
	  "use_memory_as_disk_flag": false,
	  "sub_session_timeout_ms": 30000,
	  "sub_session_hash_key": ""
	},
	"httpts": {
	  "enable": true,
	  "enable_https": true,
	  "url_pattern": "/",
	  "gop_num": 0,
	  "single_gop_max_frame_num": 0
	},
	"rtsp": {
	  "enable": {{rtspEnable}},
	  "addr": ":{{rtspPort}}",
	  "rtsps_enable": {{rtspsEnable}},
	  "rtsps_addr": ":{{rtspsPort}}",
	  "rtsps_cert_file": "./{{confPath}}/cert.pem",
	  "rtsps_key_file": "./{{confPath}}/key.pem",
	  "out_wait_key_frame_flag": true,
	  "auth_enable": false,
	  "auth_method": 1,
	  "username": "q191201771",
	  "password": "pengrl"
	},
	"record": {
	  "enable_flv": false,
	  "flv_out_path": "./{{cachePath}}/flv/",
	  "enable_mpegts": false,
	  "mpegts_out_path": "./{{cachePath}}/mpegts"
	},
	"relay_push": {
	  "enable": false,
	  "addr_list": []
	},
	"static_relay_pull": {
	  "enable": false,
	  "addr": ""
	},
	"http_api": {
	  "enable": {{httpApiEnable}},
	  "addr": ":{{httpApiPort}}"
	},
	"server_id": "1",
	"http_notify": {
		"enable": false,
		"update_interval_sec": 5,
		"on_update": "http://{{baseUrl}}:10101/on_update",
		"on_pub_start": "http://{{baseUrl}}:10101/on_pub_start",
		"on_pub_stop": "http://{{baseUrl}}:10101/on_pub_stop",
		"on_sub_start": "http://{{baseUrl}}:10101/on_sub_start",
		"on_sub_stop": "http://{{baseUrl}}:10101/on_sub_stop",
		"on_relay_pull_start": "http://{{baseUrl}}:10101/on_relay_pull_start",
		"on_relay_pull_stop": "http://{{baseUrl}}:10101/on_relay_pull_stop",
		"on_rtmp_connect": "http://{{baseUrl}}:10101/on_rtmp_connect",
		"on_server_start": "http://{{baseUrl}}:10101/on_server_start",
		"on_hls_make_ts": "http://{{baseUrl}}:10101/on_hls_make_ts"
	  },
	"simple_auth": {
	  "key": "q191201771",
	  "dangerous_lal_secret": "pengrl",
	  "pub_rtmp_enable": false,
	  "sub_rtmp_enable": false,
	  "sub_httpflv_enable": false,
	  "sub_httpts_enable": false,
	  "pub_rtsp_enable": false,
	  "sub_rtsp_enable": false,
	  "hls_m3u8_enable": false
	},
	"pprof": {
	  "enable": true,
	  "addr": ":8084"
	},
	"log": {
	  "level": 1,
	  "filename": "./logs/lalserver.log",
	  "is_to_stdout": true,
	  "is_rotate_daily": true,
	  "short_file_flag": true,
	  "timestamp_flag": true,
	  "timestamp_with_ms_flag": true,
	  "level_flag": true,
	  "assert_behavior": 1
	},
	"debug": {
	  "log_group_interval_sec": 30,
	  "log_group_max_group_num": 10,
	  "log_group_max_sub_num_per_group": 10
	}
}`
