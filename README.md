# MySQL 数据同步插件

企业内部 MySQL 数据库同步到钉钉/飞书多维表格的数据连接器插件。

## 功能特性

- 支持钉钉 AI 表格和飞书多维表格
- 可视化管理后台配置数据源
- 支持表查询和自定义 SQL 两种模式
- 字段映射与别名配置
- 用户权限管理
- 单文件部署，所有前端资源嵌入二进制

## 项目结构

```
mysql-sync-plugin/
├── backend/              # Go 后端服务
│   ├── static/           # 嵌入的前端静态文件（构建时生成）
│   └── ...
├── admin-frontend/       # 管理后台前端 (Vue 3)
├── frontend-dingtalk/    # 钉钉前端 (Vue 3)
├── frontend-feishu/      # 飞书前端 (Vue 3)
├── meta.json             # 飞书数据连接器配置
├── manifest.json         # 钉钉插件配置
├── build.ps1             # Windows 构建脚本
├── build.sh              # Linux/macOS 构建脚本
└── docs/                 # 文档
```

## 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+
- MySQL 5.7+

### 一键构建

**Windows (PowerShell):**

```powershell
# 构建 Windows 版本
.\build.ps1

# 构建 Linux 版本
.\build.ps1 -Target linux

# 构建 Linux ARM64 版本
.\build.ps1 -Target linux -Arch arm64

# 构建所有平台
.\build.ps1 -Target all

# 跳过前端构建（仅重新编译后端）
.\build.ps1 -SkipFrontend
```

**Linux/macOS (Bash):**

```bash
chmod +x build.sh

# 构建 Linux 版本
./build.sh

# 构建 Windows 版本
./build.sh -t windows

# 构建所有平台
./build.sh -t all
```

### 运行

```bash
# 创建配置文件
cat > config.json << 'EOF'
{
  "mysql": {
    "host": "127.0.0.1",
    "port": 3306,
    "database": "mysql_sync_plugin",
    "username": "root",
    "password": "your_password"
  }
}
EOF

# 运行服务
./mysql-sync-plugin-linux   # Linux
./mysql-sync-plugin.exe     # Windows
```

服务启动后访问：
- 管理后台: http://localhost:8080/admin
- 钉钉前端: http://localhost:8080/dingtalk/
- 飞书前端: http://localhost:8080/feishu/

默认管理员账号: `admin` / `admin123`

## 配置说明

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| SERVER_PORT | 服务端口 | 8080 |
| SECRET_KEY | 签名密钥 | - |
| DEBUG | 调试模式 | false |

### config.json

```json
{
  "mysql": {
    "host": "127.0.0.1",
    "port": 3306,
    "database": "mysql_sync_plugin",
    "username": "root",
    "password": "your_password"
  }
}
```

## API 接口

### 钉钉 API

- `POST /dingtalk/api/sheet_meta` - 获取表结构
- `POST /dingtalk/api/records` - 获取表记录

### 飞书 API

- `POST /feishu/api/table_meta` - 获取表结构
- `POST /feishu/api/records` - 获取表记录

### 管理后台 API

- `POST /admin/api/login` - 登录
- `GET /admin/api/datasources` - 数据源列表
- `POST /admin/api/datasources` - 创建数据源
- 更多接口请参考源码

## 技术栈

- **后端**: Go 1.21 + Gin + MySQL Driver
- **前端**: Vue 3 + TypeScript + Vite + Ant Design Vue
- **数据库**: MySQL 5.7+

## 开发

### 后端开发

```bash
cd backend
go mod download
go run main.go
```

### 前端开发

```bash
# 管理后台
cd admin-frontend
npm install
npm run dev

# 钉钉前端
cd frontend-dingtalk
npm install
npm run dev

# 飞书前端
cd frontend-feishu
npm install
npm run dev
```

## 数据库初始化

首次运行前需要初始化数据库：

```bash
mysql -u root -p < backend/schema/mysql_schema.sql
```