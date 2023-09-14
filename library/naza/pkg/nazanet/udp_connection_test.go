// Copyright 2020, Chef.  All rights reserved.
// https://github.com/q191201771/naza
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

package nazanet_test

import (
	"fmt"
	"net"
	"sync"
	"testing"

	"live-server/library/naza/pkg/assert"

	"live-server/library/naza/pkg/nazanet"
)

// [::]:4000 => 0.0.0.0:4000
// [::1]:4000 => 127.0.0.1:4000
//
// ------------------------------
// srv laddr=":4000" raddr=""
// succ:
// cli laddr=""      raddr="127.0.0.1:4000"
// cli laddr=""      raddr="[::1]:4000"
// cli laddr=":4001" raddr="127.0.0.1:4000"
// fail:
//
// ------------------------------
// srv laddr="[::]:4000" raddr=""
// succ:
// cli laddr=""          raddr="[::1]:4000"
// cli laddr=""          raddr="127.0.0.1:4000"
// fail:
//

func TestUdpConnection(t *testing.T) {
	p := nazanet.NewAvailUdpConnPool(4000, 8000)
	srvConn, srvPort, err := p.Acquire()
	assert.Equal(t, nil, err)
	toAddr1 := fmt.Sprintf("127.0.0.1:%d", srvPort)
	toAddr2 := fmt.Sprintf("[::1]:%d", srvPort)
	srv, err := nazanet.NewUdpConnection(func(option *nazanet.UdpConnectionOption) {
		option.Conn = srvConn
	})
	assert.Equal(t, nil, err)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		var count int
		err := srv.RunLoop(func(b []byte, raddr *net.UDPAddr, err error) bool {
			count++
			if count > 2 {
				return true
			}
			assert.Equal(t, []byte("hello"), b)
			err2 := srv.Write2Addr([]byte("world"), raddr)
			assert.Equal(t, nil, err2)
			return true
		})
		// 因为server loop是通过Dispose强行关闭的，所以这里error有值
		assert.IsNotNil(t, err)
	}()

	cli, err := nazanet.NewUdpConnection(func(option *nazanet.UdpConnectionOption) {
		option.RAddr = toAddr1
	})
	assert.Equal(t, nil, err)
	go func() {
		err := cli.Write([]byte("hello"))
		assert.Equal(t, nil, err)
		err = cli.RunLoop(func(b []byte, raddr *net.UDPAddr, err error) bool {
			assert.Equal(t, []byte("world"), b)
			return false
		})
		assert.Equal(t, nil, err)
		wg.Done()
	}()

	cli2, err := nazanet.NewUdpConnection(func(option *nazanet.UdpConnectionOption) {
		option.RAddr = toAddr2
	})
	assert.Equal(t, nil, err)
	go func() {
		err := cli2.Write([]byte("hello"))
		assert.Equal(t, nil, err)
		err = cli2.RunLoop(func(b []byte, raddr *net.UDPAddr, err error) bool {
			assert.Equal(t, []byte("world"), b)
			return false
		})
		assert.Equal(t, nil, err)
		wg.Done()
	}()

	wg.Wait()

	err = srv.Dispose()
	assert.Equal(t, nil, err)

	// 测试没有设置对端地址，直接使用Write的情况
	cli3, err := nazanet.NewUdpConnection()
	assert.Equal(t, nil, err)
	err = cli3.Write([]byte{'1'})
	assert.IsNotNil(t, err)
}
