package main

import (
	"context"
	"fmt"
	"live-server/internal/lalserver"
	"live-server/internal/pullrtsp2pushrtsp"
	"live-server/library/LAL/pkg/base"
	"live-server/library/naza/pkg/nazalog"
	"live-server/pkg/utils"
	"live-server/pkg/utils/network"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx      context.Context
	hide     bool // 是否不退出程序仅隐藏窗口
	isHidden bool // 是否已隐藏窗口
	isOnTop  bool // 是否已置顶窗口

	rtspTunnels map[string]struct {
		RtspTunnel *pullrtsp2pushrtsp.RtspTunnel `json:"rtsp_tunnel"`
		Infos      TunnelInfos                   `json:"infos"`
	}
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

/* *****************生命周期***************** */
// startup 启动时生命周期
func (a *App) startup(ctx context.Context) {
	// 初始化运行目录
	err := os.Chdir(utils.GetCurrPath())
	if err != nil {
		panic(err)
	}
	a.ctx = ctx
	runtime.WindowSetDarkTheme(a.ctx)
	a.rtspTunnels = make(map[string]struct {
		RtspTunnel *pullrtsp2pushrtsp.RtspTunnel `json:"rtsp_tunnel"`
		Infos      TunnelInfos                   `json:"infos"`
	})

	go lalserver.Init()
}

// beforeClose 退出前生命周期
func (a *App) beforeClose(ctx context.Context) bool {
	if a.hide {
		a.hideWindow()
	} else {
		// TODO: 退出前的操作
		fmt.Println("退出前的操作", a)
	}
	// 返回 true 将阻止程序关闭
	return a.hide
}

/* *****************http-api***************** */
func (a *App) HttpApiStatGroup(stream_name string) base.ApiStatGroupResp {
	return lalserver.HttpApiStatGroup(stream_name)
}

func (a *App) HttpApiStatAllGroup() base.ApiStatAllGroupResp {
	return lalserver.HttpApiStatAllGroup()
}

func (a *App) HttpApiCtrlStartRelayPull(url, stream_name string) base.ApiCtrlStartRelayPullResp {
	return lalserver.HttpApiCtrlStartRelayPull(url, stream_name)
}

func (a *App) HttpApiCtrlStopRelayPull(stream_name string) base.ApiCtrlStopRelayPullResp {
	return lalserver.HttpApiCtrlStopRelayPull(stream_name)
}

/* *****************pullrtsp2pushrtsp***************** */
type TunnelInfos struct {
	Name      string `json:"name"`
	OriginUrl string `json:"origin_url"`
	TargetUrl string `json:"target_url"`
	HlsUrl    string `json:"hls_url"`
}

func (a *App) PullRtsp2PushRtspStart(inUrl string, name string, pullOverTcp int, pushOverTcp int) (TunnelInfos, error) {
	// 判断是否已经存在同name
	if _, ok := a.rtspTunnels[name]; ok {
		nazalog.Errorf("tunnel already exists. name=%s", name)
		return TunnelInfos{}, fmt.Errorf("tunnel already exists. name=%s", name)
	}
	outUrl := lalserver.RtspDefaultUrl(name)

	rtspTunnel := pullrtsp2pushrtsp.NewRtspTunnel(inUrl, outUrl, pullOverTcp == 1, pushOverTcp == 1)
	infos := TunnelInfos{
		Name:      name,
		OriginUrl: inUrl,
		TargetUrl: outUrl,
		HlsUrl:    lalserver.HlsDefaultUrl(name),
	}
	if a.rtspTunnels == nil {
		a.rtspTunnels = make(map[string]struct {
			RtspTunnel *pullrtsp2pushrtsp.RtspTunnel `json:"rtsp_tunnel"`
			Infos      TunnelInfos                   `json:"infos"`
		})
	}
	a.rtspTunnels[name] = struct {
		RtspTunnel *pullrtsp2pushrtsp.RtspTunnel `json:"rtsp_tunnel"`
		Infos      TunnelInfos                   `json:"infos"`
	}{
		RtspTunnel: rtspTunnel,
		Infos:      infos,
	}

	err := rtspTunnel.Start()
	if err != nil {
		nazalog.Errorf("start tunnel failed. err=%+v", err)
		delete(a.rtspTunnels, name)
		return TunnelInfos{}, fmt.Errorf("start tunnel failed. err=%+v", err)
	}

	//go func() {
	//	time.Sleep(5 * time.Second)
	//	_ = rtspTunnel.Dispose()
	//}()

	go func() {
		err = <-rtspTunnel.WaitChan()
		nazalog.Errorf("tunnel stopped. err=%+v", err)
		delete(a.rtspTunnels, name)
		// 通知前端
		runtime.EventsEmit(a.ctx, "stream_disconnected", name)
	}()

	return infos, nil
}

// PullRtsp2PushRtspStop 停止rtsp转发任务
func (a *App) PullRtsp2PushRtspStop(name string) {
	if _, ok := a.rtspTunnels[name]; !ok {
		nazalog.Errorf("tunnel not exists. name=%s", name)
		return
	}
	err := a.rtspTunnels[name].RtspTunnel.Dispose()
	if err != nil {
		nazalog.Errorf("dispose tunnel failed. err=%+v", err)
		return
	}
	delete(a.rtspTunnels, name)
}

// PullRtsp2PushRtspStopAll 停止所有rtsp转发任务
func (a *App) PullRtsp2PushRtspStopAll() {
	for k := range a.rtspTunnels {
		err := a.rtspTunnels[k].RtspTunnel.Dispose()
		if err != nil {
			nazalog.Errorf("dispose tunnel failed. err=%+v", err)
			return
		}
		delete(a.rtspTunnels, k)
	}
}

// PullRtsp2PushRtspList 获取rtsp转发任务列表
func (a *App) PullRtsp2PushRtspList() []TunnelInfos {
	var infos []TunnelInfos
	for _, v := range a.rtspTunnels {
		infos = append(infos, v.Infos)
	}
	return infos
}

/* *****************network***************** */
func (a *App) CheckUrl(url string) (string, error) {
	return network.CheckUrl(url)
}

/* *****************私有方法***************** */
// showWindow 显示窗口
func (a *App) showWindow() {
	runtime.WindowShow(a.ctx)
	a.isHidden = false
}

// hideWindow 隐藏窗口
func (a *App) hideWindow() {
	runtime.WindowHide(a.ctx)
	a.isHidden = true
}

// toggleWindow 切换窗口显示状态
func (a *App) toggleWindow() {
	if a.isHidden {
		a.showWindow()
		return
	}
	a.hideWindow()
}

/* *****************窗口控制方法(暴露给前端)***************** */

// WindowMinimise 窗口最小化
func (a *App) WindowMinimise() {
	runtime.WindowMinimise(a.ctx)
}

// WindowMaximise 窗口最大化(切换状态)
func (a *App) WindowMaximise() {
	runtime.WindowToggleMaximise(a.ctx)
}

// WindowClose 窗口关闭(控制是否不退出仅隐藏)
func (a *App) WindowClose(isHidden bool) {
	a.hide = isHidden
	runtime.Quit(a.ctx)
}

// WindowOnTop 窗口置顶(切换状态)
func (a *App) WindowOnTop() {
	a.isOnTop = !a.isOnTop // 切换置顶flag状态
	runtime.WindowSetAlwaysOnTop(a.ctx, a.isOnTop)
}

// WindowFullScreen 窗口全屏(切换状态)
func (a *App) WindowFullScreen() {
	if !a.WindowIsFullScreen() {
		runtime.WindowFullscreen(a.ctx)
	} else {
		runtime.WindowUnfullscreen(a.ctx)
	}
}

// WindowToggle 切换窗口显示/隐藏
func (a *App) WindowShowOrHide() {
	a.toggleWindow()
}

/* *****************窗口状态(暴露给前端)***************** */

// WindowIsMaximised 窗口是否最大化
func (a *App) WindowIsMaximised() bool {
	return runtime.WindowIsMaximised(a.ctx)
}

// WindowIsMinimised 窗口是否最小化
func (a *App) WindowIsMinimised() bool {
	return runtime.WindowIsMinimised(a.ctx)
}

// WindowIsOnToped 窗口是否置顶
func (a *App) WindowIsOnToped() bool {
	return a.isOnTop
}

// WindowIsHidden 窗口是否隐藏
func (a *App) WindowIsHidden() bool {
	return a.isHidden
}

// WindowIsFullScreen 窗口是否全屏
func (a *App) WindowIsFullScreen() bool {
	return runtime.WindowIsFullscreen(a.ctx)
}

/* *****************窗口位置(暴露给前端)***************** */

// WindowSetPosition 设置窗口位置
func (a *App) WindowSetPosition(x, y int) {
	runtime.WindowSetPosition(a.ctx, x, y)
}

// WindowGetPosition 获取窗口位置
func (a *App) WindowGetPosition() (x, y int) {
	return runtime.WindowGetPosition(a.ctx)
}

// WindowCenter 设置窗口居中
func (a *App) WindowCenter() {
	runtime.WindowCenter(a.ctx)
}

/* *****************窗口大小(暴露给前端)***************** */

// WindowSetSize 设置窗口大小
func (a *App) WindowSetSize(width, height int) {
	runtime.WindowSetSize(a.ctx, width, height)
}

// WindowGetSize 获取窗口大小
func (a *App) WindowGetSize() (width, height int) {
	return runtime.WindowGetSize(a.ctx)
}
