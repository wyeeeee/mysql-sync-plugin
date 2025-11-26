# MySQL数据同步插件

企业内部MySQL数据库同步到钉钉AI表格的数据源插件。

## ✨ 功能特性

- 🔌 连接企业内部MySQL数据库
- 📊 可视化选择数据库、数据表和字段
- 🔄 自动同步数据到AI表格
- 📈 支持分页拉取大数据量
- 🔐 签名验证确保数据安全
- 🎨 友好的配置界面

## 🚀 快速开始

### 后端启动

```bash
cd backend
go mod download
go run main.go
```

### 前端启动

```bash
cd frontend
npm install
npm run dev
```

## 📖 详细文档

请查看 [部署文档](docs/部署文档.md) 了解完整的部署流程。

## 🏗️ 技术栈

- **后端**: Go 1.21 + Gin + MySQL Driver
- **前端**: React 18 + TypeScript + Vite + Ant Design
- **平台**: 钉钉开放平台酷应用

## 📂 项目结构

```
mysql-sync-plugin/
├── backend/           # Go后端服务
├── frontend/          # React前端配置页面
├── manifest.json      # 钉钉插件配置
└── docs/             # 文档
```

## 🔧 配置说明

### 后端环境变量

```bash
SERVER_PORT=8080                          # 服务端口
SECRET_KEY=your-secret-key-from-dingtalk  # 签名密钥
DEBUG=false                               # 调试模式
```

### 前端环境变量

```bash
VITE_API_BASE_URL=http://localhost:8080  # 后端API地址
```

## 📝 API接口

### 前端配置接口(无需签名)

- `POST /api/tables` - 获取数据表列表
- `POST /api/fields` - 获取表字段信息

### AI表格调用接口(需签名)

- `POST /api/sheet_meta` - 获取表结构
- `POST /api/records` - 获取表记录(分页)

## 🔐 安全要求

1. ✅ 前端必须部署到HTTPS
2. ✅ 后端实现签名验证
3. ✅ 使用MySQL只读账户
4. ✅ 不在代码中硬编码敏感信息

## 🧪 测试

```bash
# 测试后端健康检查
curl http://localhost:8080/health

# 测试获取表列表
curl -X POST http://localhost:8080/api/tables \
  -H "Content-Type: application/json" \
  -d '{"host":"127.0.0.1","port":3306,"username":"root","password":"pwd","database":"test"}'
```
