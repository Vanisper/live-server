// Copyright 2020, Chef.  All rights reserved.
// https://github.com/q191201771/naza
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

package nazahttp_test

import (
	"testing"

	"live-server/library/naza/pkg/assert"
	"live-server/library/naza/pkg/nazahttp"
)

func TestGetHttpFile(t *testing.T) {
	content, err := nazahttp.GetHttpFile("http://pengrl.com", 10000)
	assert.IsNotNil(t, content)
	assert.Equal(t, nil, err)

	content, err = nazahttp.GetHttpFile("http://127.0.0.1:12356", 10000)
	assert.Equal(t, nil, content)
	assert.IsNotNil(t, err)
}

func TestDownloadHttpFile(t *testing.T) {
	n, err := nazahttp.DownloadHttpFile("http://pengrl.com", "/tmp/index.html", 10000)
	assert.Equal(t, true, n > 0)
	assert.Equal(t, nil, err)

	n, err = nazahttp.DownloadHttpFile("http://127.0.0.1:12356", "/tmp/index.html", 10000)
	assert.IsNotNil(t, err)

	// 保存文件至不存在的本地目录下
	n, err = nazahttp.DownloadHttpFile("http://pengrl.com", "/notexist/index.html", 10000)
	assert.IsNotNil(t, err)
}
