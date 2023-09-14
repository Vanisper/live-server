// Copyright 2021, Chef.  All rights reserved.
// https://github.com/q191201771/lal
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

package base

import "live-server/library/naza/pkg/nazalog"

var Log = nazalog.GetGlobalLogger()

// AddCors2HlsFlag 是否为hls增加跨域相关的http header
var AddCors2HlsFlag = true
