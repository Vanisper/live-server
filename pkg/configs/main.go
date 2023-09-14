package configs

import (
	"embed"
	"live-server/pkg/utils"
)

var (
	ModuleName        = "live-server"
	ModuleDescription = "摄像头视频流服务"
	ModuleVer         = "v0.0.1"
	ModuleConfigPath  = utils.GetCurrPath() + "/configs"
)

// InitConfig 初始化配置
// @param confPath string 配置文件路径:此路径一定要和`go:embed all:path` 对应
func InitConfig(confPath string, conf *embed.FS) {
	// 初始化配置目录
	utils.MkDir(ModuleConfigPath)
	// 将配置文件写入到本地
	utils.DeepWalk(confPath, conf)
}
