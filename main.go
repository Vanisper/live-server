package main

import (
	"embed"
	"live-server/pkg/utils"
	"log"
	"net/http"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

type FileLoader struct {
	http.Handler
}

func NewFileLoader() *FileLoader {
	return &FileLoader{}
}

func (h *FileLoader) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	log.Println(req.URL.Path)
	requestedFilename := strings.TrimPrefix(req.URL.Path, "/") // 获取文件名称 assets-handler-x/xxxx
	log.Println(requestedFilename)
	if requestedFilename == "" {
		// 判断是否是读取index.html
		requestedFilename = "index.html"
	}
	// 静态资源基础路径设置为程序运行位置
	assetBasePath := utils.GetCurrPath() + "\\"
	// 先判断是不是assets-handler-x/xxxx  获取其中的盘符
	if strings.HasPrefix(requestedFilename, "assets-handler-") {
		// 获取盘符
		assetBasePath = requestedFilename[15:16] + "://"
		// 获取文件名称
		requestedFilename = requestedFilename[17:]
	}
	assetFullPath := assetBasePath + requestedFilename
	log.Println(assetFullPath)
	http.ServeFile(res, req, assetFullPath)
	//fileData, err := os.ReadFile( assetFullPath)
	///*
	//	读取程序的运行路径下的对应资源文件
	//	这里的资源前缀可以自定义，如果是本地资源的话就使用os.ReadFile的方式获取文件，
	//	如果是网络资源的话，可以使用 http.ServeFile() 方法获取资源文件。
	//	当然，也可以统一成http.ServeFile，就是在本地以某个文件夹起一个server即可
	//	---
	//	该方法是在 assets 中获取不到资源的情况下执行的，这种场景一般出现在：前端网页中有些网络资源可能失效了，
	//	所以在assets出现404的情况下，就会执行handler方法，用于可能的补救措施，
	//	或者说是前端有意为之的，变相地将本地资源“转发”到前端所处端口下
	//*/
	//if err != nil {
	//	println(err.Error())
	//	res.WriteHeader(http.StatusBadRequest)
	//	res.Write([]byte(fmt.Sprintf("无法加载文件 %s", requestedFilename)))
	//}
	//res.Write(fileData)
}

//go:embed build/appicon.png
var icon []byte
var isDev = false

func main() {

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:       "live-server",
		Width:       1024,
		Height:      768,
		Frameless:   true,
		StartHidden: isDev,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: NewFileLoader(),
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 0},
		OnStartup:        app.startup,
		OnBeforeClose:    app.beforeClose,
		Bind: []interface{}{
			app,
		},
		// Windows platform specific options
		Windows: &windows.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			DisableWindowIcon:    false,
			// DisableFramelessWindowDecorations: false,
			WebviewUserDataPath: "",
			ZoomFactor:          1.0,
		},
		// Mac platform specific options
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            false,
				UseToolbar:                 false,
				HideToolbarSeparator:       true,
			},
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			About: &mac.AboutInfo{
				Title:   "live-server",
				Message: "",
				Icon:    icon,
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
