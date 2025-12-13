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

// GetAdminFS 获取管理后台静态文件系统
func GetAdminFS() http.FileSystem {
	subFS, _ := fs.Sub(adminFS, "admin")
	return http.FS(subFS)
}

// GetDingtalkFS 获取钉钉前端静态文件系统
func GetDingtalkFS() http.FileSystem {
	subFS, _ := fs.Sub(dingtalkFS, "dingtalk")
	return http.FS(subFS)
}

// GetFeishuFS 获取飞书前端静态文件系统
func GetFeishuFS() http.FileSystem {
	subFS, _ := fs.Sub(feishuFS, "feishu")
	return http.FS(subFS)
}

// GetAdminEmbedFS 获取管理后台嵌入文件系统（用于读取文件内容）
func GetAdminEmbedFS() embed.FS {
	return adminFS
}

// GetDingtalkEmbedFS 获取钉钉前端嵌入文件系统
func GetDingtalkEmbedFS() embed.FS {
	return dingtalkFS
}

// GetFeishuEmbedFS 获取飞书前端嵌入文件系统
func GetFeishuEmbedFS() embed.FS {
	return feishuFS
}
