// Copyright 2020, Chef.  All rights reserved.
// https://github.com/q191201771/lal
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

package innertest

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"live-server/library/naza/pkg/nazabytes"
	"live-server/library/naza/pkg/nazalog"

	"live-server/library/LAL/pkg/httpts"

	"live-server/library/naza/pkg/filebatch"

	"live-server/library/LAL/pkg/hls"

	"live-server/library/naza/pkg/mock"

	"live-server/library/naza/pkg/nazahttp"

	"live-server/library/LAL/pkg/rtprtcp"
	"live-server/library/LAL/pkg/rtsp"
	"live-server/library/LAL/pkg/sdp"

	"live-server/library/LAL/pkg/remux"

	"live-server/library/LAL/pkg/base"

	"live-server/library/naza/pkg/nazamd5"

	"live-server/library/LAL/pkg/httpflv"
	"live-server/library/LAL/pkg/logic"
	"live-server/library/LAL/pkg/rtmp"

	"live-server/library/naza/pkg/assert"
	"live-server/library/naza/pkg/nazaatomic"
)

// 开启了一个lalserver
// rtmp pub              读取flv文件，使用rtmp协议推送至服务端
// rtmp sub, httpflv sub 分别用rtmp协议以及httpflv协议从服务端拉流，再将拉取的流保存为flv文件
// 对比三份flv文件，看是否完全一致
// hls                   并检查hls生成的m3u8和ts文件，是否和之前的完全一致

// TODO chef:
// - 加上relay push
// - 加上relay pull

var (
	t *testing.T

	mode         int // 0 正常 1 输入只有音频 2 输入只有视频
	confFilename = "../../testdata/lalserver.conf.json"
	rFlvFileName = "../../testdata/test.flv"

	pushUrl        string
	httpflvPullUrl string
	httptsPullUrl  string
	rtmpPullUrl    string
	rtspPullUrl    string

	wRtmpPullFileName     string
	wFlvPullFileName      string
	wPlaylistM3u8FileName string
	wRecordM3u8FileName   string
	wHlsTsFilePath        string
	wTsPullFileName       string

	fileTagCount          int
	httpflvPullTagCount   nazaatomic.Uint32
	rtmpPullTagCount      nazaatomic.Uint32
	httptsSize            nazaatomic.Uint32
	rtspSdpCtx            sdp.LogicContext
	rtspPullAvPacketCount nazaatomic.Uint32

	httpFlvWriter httpflv.FlvFileWriter
	rtmpWriter    httpflv.FlvFileWriter

	pushSession        *rtmp.PushSession
	httpflvPullSession *httpflv.PullSession
	rtmpPullSession    *rtmp.PullSession
	rtspPullSession    *rtsp.PullSession
)

type RtspPullObserver struct {
}

func (r RtspPullObserver) OnSdp(sdpCtx sdp.LogicContext) {
	rtspSdpCtx = sdpCtx
}

func (r RtspPullObserver) OnRtpPacket(pkt rtprtcp.RtpPacket) {
}

func (r RtspPullObserver) OnAvPacket(pkt base.AvPacket) {
	rtspPullAvPacketCount.Increment()
}

func Entry(tt *testing.T) {
	// 在MacOS只测试一次
	// 其他环境（比如github CI）上则每个package都执行，因为要生产测试覆盖率
	if runtime.GOOS == "darwin" {
		_, file, _, _ := runtime.Caller(1)
		if !strings.HasSuffix(file, "innertest_test.go") {
			return
		}
	}

	t = tt

	mode = 0
	entry()

	mode = 1
	entry()

	mode = 2
	entry()
}

func entry() {
	var entryWaitGroup sync.WaitGroup // 用于等待所有协程结束
	entryWaitGroup.Add(4)

	Log.Debugf("> innertest")

	if _, err := os.Lstat(confFilename); err != nil {
		Log.Warnf("lstat %s error. err=%+v", confFilename, err)
		return
	}
	if _, err := os.Lstat(rFlvFileName); err != nil {
		Log.Warnf("lstat %s error. err=%+v", rFlvFileName, err)
		return
	}

	httpflvPullTagCount.Store(0)
	rtmpPullTagCount.Store(0)
	httptsSize.Store(0)
	rtspPullAvPacketCount.Store(0)
	hls.Clock = mock.NewFakeClock()
	hls.Clock.Set(time.Date(2022, 1, 16, 23, 24, 25, 0, time.UTC))
	httpts.SubSessionWriteChanSize = 0

	var err error

	sm := logic.NewServerManager(func(option *logic.Option) {
		option.ConfFilename = confFilename
	})
	config := sm.Config()

	_ = os.RemoveAll(config.HlsConfig.OutPath)

	go sm.RunLoop()
	time.Sleep(100 * time.Millisecond)

	getAllHttpApi(config.HttpApiConfig.Addr)

	pushUrl = fmt.Sprintf("rtmp://127.0.0.1%s/live/innertest", config.RtmpConfig.Addr)
	httpflvPullUrl = fmt.Sprintf("http://127.0.0.1%s/live/innertest.flv", config.HttpflvConfig.HttpListenAddr)
	httptsPullUrl = fmt.Sprintf("http://127.0.0.1%s/live/innertest.ts", config.HttpflvConfig.HttpListenAddr)
	rtmpPullUrl = fmt.Sprintf("rtmp://127.0.0.1%s/live/innertest", config.RtmpConfig.Addr)
	rtspPullUrl = fmt.Sprintf("rtsp://127.0.0.1%s/live/innertest", config.RtspConfig.Addr)

	wRtmpPullFileName = "../../testdata/rtmppull.flv"
	wFlvPullFileName = "../../testdata/flvpull.flv"
	wTsPullFileName = fmt.Sprintf("../../testdata/tspull_%d.ts", mode)
	wPlaylistM3u8FileName = fmt.Sprintf("%sinnertest/playlist.m3u8", config.HlsConfig.OutPath)
	wRecordM3u8FileName = fmt.Sprintf("%sinnertest/record.m3u8", config.HlsConfig.OutPath)
	wHlsTsFilePath = fmt.Sprintf("%sinnertest/", config.HlsConfig.OutPath)

	var tags []httpflv.Tag
	originTags, err := httpflv.ReadAllTagsFromFlvFile(rFlvFileName)
	assert.Equal(t, nil, err)
	if mode == 0 {
		tags = originTags
	} else if mode == 1 {
		for _, tag := range originTags {
			if tag.Header.Type == base.RtmpTypeIdMetadata || tag.Header.Type == base.RtmpTypeIdAudio {
				tags = append(tags, tag)
			}
		}
	} else if mode == 2 {
		for _, tag := range originTags {
			if tag.Header.Type == base.RtmpTypeIdMetadata || tag.Header.Type == base.RtmpTypeIdVideo {
				tags = append(tags, tag)
			}
		}
	}
	fileTagCount = len(tags)

	err = httpFlvWriter.Open(wFlvPullFileName)
	assert.Equal(t, nil, err)
	err = httpFlvWriter.WriteRaw(httpflv.FlvHeader)
	assert.Equal(t, nil, err)

	err = rtmpWriter.Open(wRtmpPullFileName)
	assert.Equal(t, nil, err)
	err = rtmpWriter.WriteRaw(httpflv.FlvHeader)
	assert.Equal(t, nil, err)

	go func() {
		rtmpPullSession = rtmp.NewPullSession(func(option *rtmp.PullSessionOption) {
			option.ReadAvTimeoutMs = 10000
			option.ReadBufSize = 0
			option.ReuseReadMessageBufferFlag = false
		}).WithOnReadRtmpAvMsg(func(msg base.RtmpMsg) {
			tag := remux.RtmpMsg2FlvTag(msg)
			err := rtmpWriter.WriteTag(*tag)
			assert.Equal(t, nil, err)
			rtmpPullTagCount.Increment()
		})
		err := rtmpPullSession.Pull(rtmpPullUrl)
		Log.Assert(nil, err)
		err = <-rtmpPullSession.WaitChan()
		Log.Debug(err)

		entryWaitGroup.Done()
	}()

	go func() {
		var flvErr error
		httpflvPullSession = httpflv.NewPullSession(func(option *httpflv.PullSessionOption) {
			option.ReadTimeoutMs = 10000
		})
		err := httpflvPullSession.Pull(httpflvPullUrl, func(tag httpflv.Tag) {
			err := httpFlvWriter.WriteTag(tag)
			assert.Equal(t, nil, err)
			httpflvPullTagCount.Increment()
		})
		Log.Assert(nil, err)
		flvErr = <-httpflvPullSession.WaitChan()
		Log.Debug(flvErr)

		entryWaitGroup.Done()
	}()

	go func() {
		b, _ := getHttpts()
		_ = os.WriteFile(wTsPullFileName, b, 0666)
		assert.Equal(t, goldenHttptsLenList[mode], len(b))
		assert.Equal(t, goldenHttptsMd5List[mode], nazamd5.Md5(b))

		entryWaitGroup.Done()
	}()
	time.Sleep(100 * time.Millisecond)

	// TODO(chef): [test] rtsp sub没有验证收到的数据，因为即使是先sub，它还有一个数据到来后，才能完成信令交互的逻辑 202206
	// TODO(chef): [perf] [2021.12.25] rtmp推rtsp拉的性能。开启rtsp pull后，rtmp pull的总时长增加了
	go func() {
		var rtspPullObserver RtspPullObserver
		rtspPullSession = rtsp.NewPullSession(&rtspPullObserver, func(option *rtsp.PullSessionOption) {
			option.PullTimeoutMs = 10000
		})
		err := rtspPullSession.Pull(rtspPullUrl)
		assert.Equal(t, nil, err)
		entryWaitGroup.Done()
	}()

	time.Sleep(100 * time.Millisecond)

	pushSession = rtmp.NewPushSession(func(option *rtmp.PushSessionOption) {
		option.WriteBufSize = 4096
		//option.WriteChanSize = 1024
	})
	err = pushSession.Push(pushUrl)
	assert.Equal(t, nil, err)

	for _, tag := range tags {
		assert.Equal(t, nil, err)
		chunks := remux.FlvTag2RtmpChunks(tag)
		//Log.Debugf("rtmp push: %d", fileTagCount.Load())
		err := pushSession.Write(chunks)
		assert.Equal(t, nil, err)
	}
	err = pushSession.Flush()
	assert.Equal(t, nil, err)

	getAllHttpApi(config.HttpApiConfig.Addr)

	// 注意，先释放push，触发pub释放，从而刷新hls的结束时切片逻辑
	pushSession.Dispose()

	for {
		if httpflvPullTagCount.Load() == uint32(fileTagCount) &&
			rtmpPullTagCount.Load() == uint32(fileTagCount) &&
			httptsSize.Load() == uint32(goldenHttptsLenList[mode]) {
			break
		}
		nazalog.Debugf("%d(%d, %d) %d(%d)",
			fileTagCount, httpflvPullTagCount.Load(), rtmpPullTagCount.Load(),
			goldenHttptsLenList[mode], httptsSize.Load())
		time.Sleep(100 * time.Millisecond)
	}

	Log.Debug("[innertest] start dispose.")

	httpflvPullSession.Dispose()
	rtmpPullSession.Dispose()
	rtspPullSession.Dispose()

	httpFlvWriter.Dispose()
	rtmpWriter.Dispose()

	// 由于windows没有信号，会导致编译错误，所以直接调用Dispose
	//_ = syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)
	sm.Dispose()

	entryWaitGroup.Wait()

	Log.Debugf("tag count. in=%d, out httpflv=%d, out rtmp=%d, out rtsp=%d",
		fileTagCount, httpflvPullTagCount.Load(), rtmpPullTagCount.Load(), rtspPullAvPacketCount.Load())

	compareFile()
	goldenRtspSdpTmplList[mode] = strings.ReplaceAll(goldenRtspSdpTmplList[mode], "{atoolv}", base.LalPackSdp)
	assert.Equal(t, strings.ReplaceAll(goldenRtspSdpTmplList[mode], "\n", "\r\n"), string(rtspSdpCtx.RawSdp))
}

func compareFile() {
	r, err := os.ReadFile(rFlvFileName)
	assert.Equal(t, nil, err)
	Log.Debugf("%s filesize:%d", rFlvFileName, len(r))

	// 检查httpflv
	w, err := os.ReadFile(wFlvPullFileName)
	assert.Equal(t, nil, err)
	assert.Equal(t, goldenHttpflvLenList[mode], len(w))
	assert.Equal(t, goldenHttpflvMd5List[mode], nazamd5.Md5(w))

	// 检查rtmp
	w, err = os.ReadFile(wRtmpPullFileName)
	assert.Equal(t, nil, err)
	assert.Equal(t, goldenRtmpLenList[mode], len(w))
	assert.Equal(t, goldenRtmpMd5List[mode], nazamd5.Md5(w))

	// 检查hls的m3u8文件
	playListM3u8, err := os.ReadFile(wPlaylistM3u8FileName)
	assert.Equal(t, nil, err)
	assert.Equal(t, goldenPlaylistM3u8List[mode], string(playListM3u8))
	recordM3u8, err := os.ReadFile(wRecordM3u8FileName)
	assert.Equal(t, nil, err)
	assert.Equal(t, goldenRecordM3u8List[mode], string(recordM3u8))

	// 检查hls的ts文件
	var allContent []byte
	var fileNum int
	err = filebatch.Walk(
		wHlsTsFilePath,
		false,
		".ts",
		func(path string, info os.FileInfo, content []byte, err error) []byte {
			allContent = append(allContent, content...)
			fileNum++
			return nil
		})
	assert.Equal(t, nil, err)
	allContentMd5 := nazamd5.Md5(allContent)
	assert.Equal(t, goldenHlsTsNumList[mode], fileNum)
	assert.Equal(t, goldenHlsTsLenList[mode], len(allContent))
	assert.Equal(t, goldenHlsTsMd5List[mode], allContentMd5)
}

func getAllHttpApi(addr string) {
	var b []byte
	var err error

	b, err = httpGet(fmt.Sprintf("http://%s/api/stat/lal_info", addr))
	Log.Assert(nil, err)
	Log.Debugf("%s", string(b))

	b, err = httpGet(fmt.Sprintf("http://%s/api/stat/group?stream_name=innertest", addr))
	Log.Assert(nil, err)
	Log.Debugf("%s", string(b))

	b, err = httpGet(fmt.Sprintf("http://%s/api/stat/all_group", addr))
	Log.Assert(nil, err)
	Log.Debugf("%s", string(b))

	var acspr base.ApiCtrlStartRelayPullReq
	b, err = httpPost(fmt.Sprintf("http://%s/api/ctrl/start_relay_pull", addr), &acspr)
	Log.Assert(nil, err)
	Log.Debugf("%s", string(b))

	var ackos base.ApiCtrlKickSessionReq
	b, err = httpPost(fmt.Sprintf("http://%s/api/ctrl/kick_session", addr), &ackos)
	Log.Assert(nil, err)
	Log.Debugf("%s", string(b))
}

func getHttpts() ([]byte, error) {
	resp, err := http.DefaultClient.Get(httptsPullUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var buf nazabytes.Buffer
	buf.ReserveBytes(goldenHttptsLenList[mode])
	for {
		n, err := resp.Body.Read(buf.WritableBytes())
		if n > 0 {
			buf.Flush(n)
			httptsSize.Add(uint32(n))
		}
		if err != nil {
			return buf.Bytes(), err
		}
		if buf.Len() == goldenHttptsLenList[mode] {
			return buf.Bytes(), nil
		}
	}
}

// ---------------------------------------------------------------------------------------------------------------------

// TODO(chef): refactor 移入naza中

func httpGet(url string) ([]byte, error) {
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func httpPost(url string, info interface{}) ([]byte, error) {
	resp, err := nazahttp.PostJson(url, info, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// ---------------------------------------------------------------------------------------------------------------------

var (
	goldenRtmpLenList = []int{2120047, 504722, 1615715}
	goldenRtmpMd5List = []string{
		"7d68f0e2ab85c1992f70740479c8d3db",
		"b889f690e07399c8c8353a3b1dba7efb",
		"b5a9759455039761b6d4dd3ed8e97634",
	}

	goldenHttpflvLenList = []int{2120047, 504722, 1615715}
	goldenHttpflvMd5List = []string{
		"7d68f0e2ab85c1992f70740479c8d3db",
		"b889f690e07399c8c8353a3b1dba7efb",
		"b5a9759455039761b6d4dd3ed8e97634",
	}

	goldenHlsTsNumList = []int{8, 10, 8}
	goldenHlsTsLenList = []int{2219152, 525648, 1696512}
	goldenHlsTsMd5List = []string{
		"5a05b84486382f3b8d2a0f9e75c67623",
		"f03c5ab24dddd8875f06cdea605cd87c",
		"ceec699eae6671507376c26ac1cdaba4",
	}

	goldenHttptsLenList = []int{2216332, 522264, 1693880}
	goldenHttptsMd5List = []string{
		"3ba4baa20df968196eac75d96f8041b5",
		"0ec316060f1aeeb1d283b30158c2eeb8",
		"e0d90dd5efd1119f1f66db8ac4cf5d48",
	}
)

var goldenPlaylistM3u8List = []string{
	`#EXTM3U
#EXT-X-VERSION:3
#EXT-X-ALLOW-CACHE:NO
#EXT-X-TARGETDURATION:5
#EXT-X-MEDIA-SEQUENCE:2

#EXTINF:3.333,
innertest-1642375465000-2.ts
#EXTINF:4.000,
innertest-1642375465000-3.ts
#EXTINF:4.867,
innertest-1642375465000-4.ts
#EXTINF:3.133,
innertest-1642375465000-5.ts
#EXTINF:4.000,
innertest-1642375465000-6.ts
#EXTINF:2.644,
innertest-1642375465000-7.ts
#EXT-X-ENDLIST
`,
	`#EXTM3U
#EXT-X-VERSION:3
#EXT-X-ALLOW-CACHE:NO
#EXT-X-TARGETDURATION:3
#EXT-X-MEDIA-SEQUENCE:4

#EXTINF:3.088,
innertest-1642375465000-4.ts
#EXTINF:3.088,
innertest-1642375465000-5.ts
#EXTINF:3.089,
innertest-1642375465000-6.ts
#EXTINF:3.088,
innertest-1642375465000-7.ts
#EXTINF:3.088,
innertest-1642375465000-8.ts
#EXTINF:2.113,
innertest-1642375465000-9.ts
#EXT-X-ENDLIST
`,
	`#EXTM3U
#EXT-X-VERSION:3
#EXT-X-ALLOW-CACHE:NO
#EXT-X-TARGETDURATION:5
#EXT-X-MEDIA-SEQUENCE:2

#EXTINF:3.333,
innertest-1642375465000-2.ts
#EXTINF:4.000,
innertest-1642375465000-3.ts
#EXTINF:4.867,
innertest-1642375465000-4.ts
#EXTINF:3.133,
innertest-1642375465000-5.ts
#EXTINF:4.000,
innertest-1642375465000-6.ts
#EXTINF:2.600,
innertest-1642375465000-7.ts
#EXT-X-ENDLIST
`,
}

var goldenRecordM3u8List = []string{
	`#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:5
#EXT-X-MEDIA-SEQUENCE:0

#EXT-X-DISCONTINUITY
#EXTINF:4.000,
innertest-1642375465000-0.ts
#EXTINF:4.000,
innertest-1642375465000-1.ts
#EXTINF:3.333,
innertest-1642375465000-2.ts
#EXTINF:4.000,
innertest-1642375465000-3.ts
#EXTINF:4.867,
innertest-1642375465000-4.ts
#EXTINF:3.133,
innertest-1642375465000-5.ts
#EXTINF:4.000,
innertest-1642375465000-6.ts
#EXTINF:2.644,
innertest-1642375465000-7.ts
#EXT-X-ENDLIST
`,
	`#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:3
#EXT-X-MEDIA-SEQUENCE:0

#EXT-X-DISCONTINUITY
#EXTINF:3.088,
innertest-1642375465000-0.ts
#EXTINF:3.088,
innertest-1642375465000-1.ts
#EXTINF:3.089,
innertest-1642375465000-2.ts
#EXTINF:3.088,
innertest-1642375465000-3.ts
#EXTINF:3.088,
innertest-1642375465000-4.ts
#EXTINF:3.088,
innertest-1642375465000-5.ts
#EXTINF:3.089,
innertest-1642375465000-6.ts
#EXTINF:3.088,
innertest-1642375465000-7.ts
#EXTINF:3.088,
innertest-1642375465000-8.ts
#EXTINF:2.113,
innertest-1642375465000-9.ts
#EXT-X-ENDLIST
`,
	`#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:5
#EXT-X-MEDIA-SEQUENCE:0

#EXT-X-DISCONTINUITY
#EXTINF:4.000,
innertest-1642375465000-0.ts
#EXTINF:4.000,
innertest-1642375465000-1.ts
#EXTINF:3.333,
innertest-1642375465000-2.ts
#EXTINF:4.000,
innertest-1642375465000-3.ts
#EXTINF:4.867,
innertest-1642375465000-4.ts
#EXTINF:3.133,
innertest-1642375465000-5.ts
#EXTINF:4.000,
innertest-1642375465000-6.ts
#EXTINF:2.600,
innertest-1642375465000-7.ts
#EXT-X-ENDLIST
`,
}

var goldenRtspSdpTmplList = []string{
	`v=0
o=- 0 0 IN IP4 127.0.0.1
s=No Name
c=IN IP4 127.0.0.1
t=0 0
a=tool:{atoolv}
m=video 0 RTP/AVP 96
a=rtpmap:96 H264/90000
a=fmtp:96 packetization-mode=1; sprop-parameter-sets=Z2QAFqyyAUBf8uAiAAADAAIAAAMAPB4sXJA=,aOvDyyLA; profile-level-id=640016
a=control:streamid=0
m=audio 0 RTP/AVP 97
b=AS:128
a=rtpmap:97 MPEG4-GENERIC/44100/2
a=fmtp:97 profile-level-id=1;mode=AAC-hbr;sizelength=13;indexlength=3;indexdeltalength=3; config=121056e500
a=control:streamid=1
`,
	`v=0
o=- 0 0 IN IP4 127.0.0.1
s=No Name
c=IN IP4 127.0.0.1
t=0 0
a=tool:{atoolv}
m=audio 0 RTP/AVP 97
b=AS:128
a=rtpmap:97 MPEG4-GENERIC/44100/2
a=fmtp:97 profile-level-id=1;mode=AAC-hbr;sizelength=13;indexlength=3;indexdeltalength=3; config=121056e500
a=control:streamid=0
`,
	`v=0
o=- 0 0 IN IP4 127.0.0.1
s=No Name
c=IN IP4 127.0.0.1
t=0 0
a=tool:{atoolv}
m=video 0 RTP/AVP 96
a=rtpmap:96 H264/90000
a=fmtp:96 packetization-mode=1; sprop-parameter-sets=Z2QAFqyyAUBf8uAiAAADAAIAAAMAPB4sXJA=,aOvDyyLA; profile-level-id=640016
a=control:streamid=0
`,
}
