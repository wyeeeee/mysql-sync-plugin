#!/bin/bash
# MySQL Sync Plugin 一键构建脚本
# 编译所有前端并嵌入到单个 Go 二进制文件
# 支持交叉编译 Windows/Linux/macOS

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
GRAY='\033[0;37m'
NC='\033[0m'

# 默认参数
OUTPUT=""
TARGET="linux"
ARCH="amd64"
SKIP_FRONTEND=false

# 帮助信息
show_help() {
    echo "MySQL Sync Plugin 构建脚本"
    echo ""
    echo "用法: ./build.sh [选项]"
    echo ""
    echo "选项:"
    echo "  -t, --target <平台>    目标平台: windows, linux, darwin, all (默认: linux)"
    echo "  -a, --arch <架构>      目标架构: amd64, arm64 (默认: amd64)"
    echo "  -o, --output <文件名>  输出文件名 (默认: 自动生成)"
    echo "  --skip-frontend        跳过前端构建"
    echo "  -h, --help             显示帮助信息"
    echo ""
    echo "示例:"
    echo "  ./build.sh                          # 构建 Linux amd64"
    echo "  ./build.sh -t windows               # 构建 Windows amd64"
    echo "  ./build.sh -t linux -a arm64        # 构建 Linux arm64"
    echo "  ./build.sh -t all                   # 构建所有平台"
    echo "  ./build.sh --skip-frontend          # 跳过前端构建"
}

# 解析参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--target)
            TARGET="$2"
            shift 2
            ;;
        -a|--arch)
            ARCH="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT="$2"
            shift 2
            ;;
        --skip-frontend)
            SKIP_FRONTEND=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

# 根据目标平台生成输出文件名
get_output_name() {
    local os=$1
    local arch=$2
    local name="mysql-sync-plugin"
    if [ "$arch" != "amd64" ]; then
        name="${name}-${arch}"
    fi
    if [ "$os" = "windows" ]; then
        echo "${name}.exe"
    else
        echo "${name}-${os}"
    fi
}

PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"

echo -e "${CYAN}========================================"
echo -e "  MySQL Sync Plugin 构建脚本"
echo -e "========================================${NC}"
echo ""

# 检查 Node.js
if ! command -v npm &> /dev/null; then
    echo -e "${RED}[错误] 未找到 npm，请先安装 Node.js${NC}"
    exit 1
fi

# 检查 Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}[错误] 未找到 go，请先安装 Go${NC}"
    exit 1
fi

# 定义路径
ADMIN_FRONTEND="$PROJECT_ROOT/admin-frontend"
DINGTALK_FRONTEND="$PROJECT_ROOT/frontend-dingtalk"
FEISHU_FRONTEND="$PROJECT_ROOT/frontend-feishu"
BACKEND="$PROJECT_ROOT/backend"
STATIC_DIR="$BACKEND/static"

# 清理旧的静态文件
echo -e "${YELLOW}[1/6] 清理旧的静态文件...${NC}"
rm -rf "$STATIC_DIR/admin" "$STATIC_DIR/dingtalk" "$STATIC_DIR/feishu"
mkdir -p "$STATIC_DIR/admin" "$STATIC_DIR/dingtalk" "$STATIC_DIR/feishu"

if [ "$SKIP_FRONTEND" = false ]; then
    # 构建管理后台前端
    echo -e "${YELLOW}[2/6] 构建管理后台前端...${NC}"
    cd "$ADMIN_FRONTEND"
    if [ ! -d "node_modules" ]; then
        echo -e "${GRAY}  安装依赖...${NC}"
        npm install --silent
    fi
    npm run build --silent

    # 构建钉钉前端
    echo -e "${YELLOW}[3/6] 构建钉钉前端...${NC}"
    cd "$DINGTALK_FRONTEND"
    if [ ! -d "node_modules" ]; then
        echo -e "${GRAY}  安装依赖...${NC}"
        npm install --silent
    fi
    npm run build --silent

    # 构建飞书前端
    echo -e "${YELLOW}[4/6] 构建飞书前端...${NC}"
    cd "$FEISHU_FRONTEND"
    if [ ! -d "node_modules" ]; then
        echo -e "${GRAY}  安装依赖...${NC}"
        npm install --silent
    fi
    npm run build --silent
else
    echo -e "${GRAY}[2-4/6] 跳过前端构建...${NC}"
fi

# 复制前端构建产物到 static 目录
echo -e "${YELLOW}[5/6] 复制前端构建产物...${NC}"

# 管理后台
if [ -d "$ADMIN_FRONTEND/dist" ]; then
    cp -r "$ADMIN_FRONTEND/dist/"* "$STATIC_DIR/admin/"
    echo -e "${GRAY}  管理后台: $(find "$STATIC_DIR/admin" -type f | wc -l) 个文件${NC}"
else
    echo -e "${YELLOW}  [警告] 管理后台构建产物不存在${NC}"
fi

# 钉钉前端
if [ -d "$DINGTALK_FRONTEND/dist" ]; then
    cp -r "$DINGTALK_FRONTEND/dist/"* "$STATIC_DIR/dingtalk/"
    echo -e "${GRAY}  钉钉前端: $(find "$STATIC_DIR/dingtalk" -type f | wc -l) 个文件${NC}"
else
    echo -e "${YELLOW}  [警告] 钉钉前端构建产物不存在${NC}"
fi

# 飞书前端
if [ -d "$FEISHU_FRONTEND/dist" ]; then
    cp -r "$FEISHU_FRONTEND/dist/"* "$STATIC_DIR/feishu/"
    # 复制 meta.json 到飞书静态目录
    if [ -f "$PROJECT_ROOT/meta.json" ]; then
        cp "$PROJECT_ROOT/meta.json" "$STATIC_DIR/feishu/"
    fi
    echo -e "${GRAY}  飞书前端: $(find "$STATIC_DIR/feishu" -type f | wc -l) 个文件${NC}"
else
    echo -e "${YELLOW}  [警告] 飞书前端构建产物不存在${NC}"
fi

# 构建 Go 二进制
echo -e "${YELLOW}[6/6] 构建 Go 二进制...${NC}"
cd "$BACKEND"

LDFLAGS="-s -w"

# 定义要构建的目标平台
if [ "$TARGET" = "all" ]; then
    TARGETS="windows:amd64 linux:amd64 linux:arm64 darwin:amd64 darwin:arm64"
else
    TARGETS="$TARGET:$ARCH"
fi

for target in $TARGETS; do
    os="${target%:*}"
    arch="${target#*:}"

    # 设置输出文件名
    if [ -n "$OUTPUT" ] && [ "$TARGET" != "all" ]; then
        output_name="$OUTPUT"
    else
        output_name=$(get_output_name "$os" "$arch")
    fi
    OUTPUT_PATH="$PROJECT_ROOT/$output_name"

    echo -e "${GRAY}  构建 $os/$arch -> $output_name${NC}"

    # 设置交叉编译环境变量
    export GOOS="$os"
    export GOARCH="$arch"
    export CGO_ENABLED=0

    go build -ldflags "$LDFLAGS" -o "$OUTPUT_PATH" .

    FILE_SIZE=$(du -h "$OUTPUT_PATH" | cut -f1)
    echo -e "${GRAY}    文件大小: $FILE_SIZE${NC}"
done

# 恢复环境变量
unset GOOS GOARCH CGO_ENABLED

echo ""
echo -e "${GREEN}========================================"
echo -e "  构建完成!"
echo -e "========================================${NC}"
echo ""
echo -e "${CYAN}使用示例:${NC}"
echo -e "${GRAY}  ./build.sh                          # 构建 Linux amd64${NC}"
echo -e "${GRAY}  ./build.sh -t windows               # 构建 Windows amd64${NC}"
echo -e "${GRAY}  ./build.sh -t linux -a arm64        # 构建 Linux arm64${NC}"
echo -e "${GRAY}  ./build.sh -t all                   # 构建所有平台${NC}"
echo -e "${GRAY}  ./build.sh --skip-frontend          # 跳过前端构建${NC}"
echo ""
