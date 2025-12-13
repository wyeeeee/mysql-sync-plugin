package static

import (
	"embed"
	"io/fs"
	"net/http"
)

// 嵌入前端静态文件
// 构建前需要先将前端编译产物复制到对应目录

//go:embed all:admin
var adminFS embed.FS

//go:embed all:dingtalk
var dingtalkFS embed.FS

//go:embed all:feishu
var feishuFS embed.FS

// GetAdminAssetsFS 获取管理后台 assets 文件系统
func GetAdminAssetsFS() http.FileSystem {
	subFS, _ := fs.Sub(adminFS, "admin/assets")
	return http.FS(subFS)
}

// GetDingtalkAssetsFS 获取钉钉前端 assets 文件系统
func GetDingtalkAssetsFS() http.FileSystem {
	subFS, _ := fs.Sub(dingtalkFS, "dingtalk/assets")
	return http.FS(subFS)
}

// GetFeishuAssetsFS 获取飞书前端 assets 文件系统
func GetFeishuAssetsFS() http.FileSystem {
	subFS, _ := fs.Sub(feishuFS, "feishu/assets")
	return http.FS(subFS)
}

// GetAdminIndexHTML 获取管理后台 index.html 内容
func GetAdminIndexHTML() ([]byte, error) {
	return adminFS.ReadFile("admin/index.html")
}

// GetDingtalkIndexHTML 获取钉钉前端 index.html 内容
func GetDingtalkIndexHTML() ([]byte, error) {
	return dingtalkFS.ReadFile("dingtalk/index.html")
}

// GetFeishuIndexHTML 获取飞书前端 index.html 内容
func GetFeishuIndexHTML() ([]byte, error) {
	return feishuFS.ReadFile("feishu/index.html")
}

// GetDingtalkFavicon 获取钉钉前端 favicon.ico 内容
func GetDingtalkFavicon() ([]byte, error) {
	return dingtalkFS.ReadFile("dingtalk/favicon.ico")
}

// GetFeishuFavicon 获取飞书前端 favicon.ico 内容
func GetFeishuFavicon() ([]byte, error) {
	return feishuFS.ReadFile("feishu/favicon.ico")
}

// GetFeishuMetaJSON 获取飞书 meta.json 内容
func GetFeishuMetaJSON() ([]byte, error) {
	return feishuFS.ReadFile("feishu/meta.json")
}
