// Copyright 2022, Chef.  All rights reserved.
// https://github.com/q191201771/lal
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

package innertest

import (
	"io"
	"live-server/library/LAL/pkg/base"
	"live-server/library/LAL/pkg/httpflv"
	"live-server/library/LAL/pkg/remux"
	"live-server/library/LAL/pkg/rtprtcp"
	"live-server/library/LAL/pkg/rtsp"
	"live-server/library/LAL/pkg/sdp"
	"testing"

	"live-server/library/naza/pkg/nazalog"
)

// TestDump_Rtsp
//
// 重放业务方的rtsp流。
//
// 本测试函数模拟客户端，读取业务方对的dumpfile，解析为rtp，合帧，写flv文件
//
// 步骤：
//
// 1. 让业务方提供lalserver录制下来的dumpfile文件
// 2. 将dumpfile存放在下面filename变量处，或者修改下面filename变量值
// 3. 执行该测试
// go test -test.run TestDump_Rtsp
func TestDump_Rtsp(t *testing.T) {
	// TODO(chef): [test] 合帧测试，只有音频部分，没有视频部分 202211

	filename := "/tmp/outpullrtsp.laldump"
	outFlvFilename := "/tmp/outtestdumprtsp.flv"

	// 初始化输出的flv文件
	var fileWriter httpflv.FlvFileWriter
	err := fileWriter.Open(outFlvFilename)
	nazalog.Assert(nil, err)
	defer fileWriter.Dispose()
	err = fileWriter.WriteRaw(httpflv.FlvHeader)
	nazalog.Assert(nil, err)

	// 初始化remuxer
	remuxer := remux.NewAvPacket2RtmpRemuxer().WithOnRtmpMsg(func(msg base.RtmpMsg) {
		nazalog.Debugf("< remuxer. %s", msg.DebugString())
		err = fileWriter.WriteTag(*remux.RtmpMsg2FlvTag(msg))
		nazalog.Assert(nil, err)
	})

	var ctx sdp.LogicContext
	var unpacker rtprtcp.IRtpUnpacker
	var unpackerVideo rtprtcp.IRtpUnpacker
	var q *rtsp.AvPacketQueue

	df := base.NewDumpFile()
	err = df.OpenToRead(filename)
	nazalog.Assert(nil, err)

	if rtsp.BaseInSessionTimestampFilterFlag {
		q = rtsp.NewAvPacketQueue(func(pkt base.AvPacket) {
			remuxer.FeedAvPacket(pkt)
		})
	}

	for {
		m, err := df.ReadOneMessage()
		nazalog.Debugf("< ReadOneMessage. %+v, %+v", m, err)
		if err == io.EOF {
			return
		}
		nazalog.Assert(nil, err)

		if m.Typ == base.DumpTypeInnerFileHeaderData {
			continue
		}

		if m.Typ != base.DumpTypeRtspRtpData && m.Typ != base.DumpTypeRtspSdpData {
			nazalog.Errorf("unknown type. typ=%d", m.Typ)
			return
		}

		if m.Typ == base.DumpTypeRtspSdpData {
			ctx, err = sdp.ParseSdp2LogicContext([]byte(m.Body))
			nazalog.Debugf("parse sdp, %+v, %+v", ctx, err)

			remuxer.OnSdp(ctx)
			unpacker = rtprtcp.DefaultRtpUnpackerFactory(ctx.GetAudioPayloadTypeBase(), ctx.AudioClockRate, 1024, func(pkt base.AvPacket) {
				nazalog.Debugf("audio avpacket. %s", pkt.DebugString())
				if rtsp.BaseInSessionTimestampFilterFlag {
					q.Feed(pkt)
				} else {
					remuxer.OnAvPacket(pkt)
				}
			})
			unpackerVideo = rtprtcp.DefaultRtpUnpackerFactory(ctx.GetVideoPayloadTypeBase(), ctx.VideoClockRate, 1024, func(pkt base.AvPacket) {
				nazalog.Debugf("video avpacket. %s", pkt.DebugString())
				if rtsp.BaseInSessionTimestampFilterFlag {
					q.Feed(pkt)
				} else {
					remuxer.OnAvPacket(pkt)
				}
			})
			continue
		}

		pkt, err := rtprtcp.ParseRtpPacket(m.Body)
		nazalog.Assert(nil, err)
		nazalog.Debugf("< ParseRtpPacket. %s", pkt.DebugString())
		if ctx.IsAudioPayloadTypeOrigin(int(pkt.Header.PacketType)) {
			unpacker.Feed(pkt)
		} else if ctx.IsVideoPayloadTypeOrigin(int(pkt.Header.PacketType)) {
			unpackerVideo.Feed(pkt)
		} else {
			nazalog.Errorf("unknown payload type. pt=%d", pkt.Header.PacketType)
		}
	}
}
