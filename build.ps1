# MySQL Sync Plugin 一键构建脚本
# 编译所有前端并嵌入到单个 Go 二进制文件
# 支持交叉编译 Windows/Linux/macOS

param(
    [string]$Output = "",
    [ValidateSet("windows", "linux", "darwin", "all")]
    [string]$Target = "windows",
    [ValidateSet("amd64", "arm64")]
    [string]$Arch = "amd64",
    [switch]$SkipFrontend
)

$ErrorActionPreference = "Stop"
$ProjectRoot = $PSScriptRoot

# 根据目标平台设置默认输出文件名
function Get-OutputName($os, $arch) {
    $name = "mysql-sync-plugin"
    if ($arch -ne "amd64") { $name += "-$arch" }
    if ($os -eq "windows") { return "$name.exe" }
    return "$name-$os"
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  MySQL Sync Plugin 构建脚本" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 检查 Node.js
if (-not (Get-Command "npm" -ErrorAction SilentlyContinue)) {
    Write-Host "[错误] 未找到 npm，请先安装 Node.js" -ForegroundColor Red
    exit 1
}

# 检查 Go
if (-not (Get-Command "go" -ErrorAction SilentlyContinue)) {
    Write-Host "[错误] 未找到 go，请先安装 Go" -ForegroundColor Red
    exit 1
}

# 定义路径
$AdminFrontend = Join-Path $ProjectRoot "admin-frontend"
$DingtalkFrontend = Join-Path $ProjectRoot "frontend-dingtalk"
$FeishuFrontend = Join-Path $ProjectRoot "frontend-feishu"
$Backend = Join-Path $ProjectRoot "backend"
$StaticDir = Join-Path $Backend "static"

# 清理旧的静态文件
Write-Host "[1/6] 清理旧的静态文件..." -ForegroundColor Yellow
$AdminStatic = Join-Path $StaticDir "admin"
$DingtalkStatic = Join-Path $StaticDir "dingtalk"
$FeishuStatic = Join-Path $StaticDir "feishu"

if (Test-Path $AdminStatic) { Remove-Item -Recurse -Force $AdminStatic }
if (Test-Path $DingtalkStatic) { Remove-Item -Recurse -Force $DingtalkStatic }
if (Test-Path $FeishuStatic) { Remove-Item -Recurse -Force $FeishuStatic }

New-Item -ItemType Directory -Force -Path $AdminStatic | Out-Null
New-Item -ItemType Directory -Force -Path $DingtalkStatic | Out-Null
New-Item -ItemType Directory -Force -Path $FeishuStatic | Out-Null

if (-not $SkipFrontend) {
    # 构建管理后台前端
    Write-Host "[2/6] 构建管理后台前端..." -ForegroundColor Yellow
    Push-Location $AdminFrontend
    try {
        if (-not (Test-Path "node_modules")) {
            Write-Host "  安装依赖..." -ForegroundColor Gray
            npm install --silent
        }
        npm run build --silent
        if ($LASTEXITCODE -ne 0) { throw "管理后台前端构建失败" }
    } finally {
        Pop-Location
    }

    # 构建钉钉前端
    Write-Host "[3/6] 构建钉钉前端..." -ForegroundColor Yellow
    Push-Location $DingtalkFrontend
    try {
        if (-not (Test-Path "node_modules")) {
            Write-Host "  安装依赖..." -ForegroundColor Gray
            npm install --silent
        }
        npm run build --silent
        if ($LASTEXITCODE -ne 0) { throw "钉钉前端构建失败" }
    } finally {
        Pop-Location
    }

    # 构建飞书前端
    Write-Host "[4/6] 构建飞书前端..." -ForegroundColor Yellow
    Push-Location $FeishuFrontend
    try {
        if (-not (Test-Path "node_modules")) {
            Write-Host "  安装依赖..." -ForegroundColor Gray
            npm install --silent
        }
        npm run build --silent
        if ($LASTEXITCODE -ne 0) { throw "飞书前端构建失败" }
    } finally {
        Pop-Location
    }
} else {
    Write-Host "[2-4/6] 跳过前端构建..." -ForegroundColor Gray
}

# 复制前端构建产物到 static 目录
Write-Host "[5/6] 复制前端构建产物..." -ForegroundColor Yellow

# 管理后台
$AdminDist = Join-Path $AdminFrontend "dist"
if (Test-Path $AdminDist) {
    Copy-Item -Recurse -Force "$AdminDist\*" $AdminStatic
    Write-Host "  管理后台: $((Get-ChildItem -Recurse $AdminStatic).Count) 个文件" -ForegroundColor Gray
} else {
    Write-Host "  [警告] 管理后台构建产物不存在" -ForegroundColor Yellow
}

# 钉钉前端
$DingtalkDist = Join-Path $DingtalkFrontend "dist"
if (Test-Path $DingtalkDist) {
    Copy-Item -Recurse -Force "$DingtalkDist\*" $DingtalkStatic
    Write-Host "  钉钉前端: $((Get-ChildItem -Recurse $DingtalkStatic).Count) 个文件" -ForegroundColor Gray
} else {
    Write-Host "  [警告] 钉钉前端构建产物不存在" -ForegroundColor Yellow
}

# 飞书前端
$FeishuDist = Join-Path $FeishuFrontend "dist"
if (Test-Path $FeishuDist) {
    Copy-Item -Recurse -Force "$FeishuDist\*" $FeishuStatic
    # 复制 meta.json 到飞书静态目录
    $MetaJson = Join-Path $ProjectRoot "meta.json"
    if (Test-Path $MetaJson) {
        Copy-Item -Force $MetaJson $FeishuStatic
    }
    Write-Host "  飞书前端: $((Get-ChildItem -Recurse $FeishuStatic).Count) 个文件" -ForegroundColor Gray
} else {
    Write-Host "  [警告] 飞书前端构建产物不存在" -ForegroundColor Yellow
}

# 构建 Go 二进制
Write-Host "[6/6] 构建 Go 二进制..." -ForegroundColor Yellow

# 定义要构建的目标平台
$targets = @()
if ($Target -eq "all") {
    $targets = @(
        @{ OS = "windows"; Arch = "amd64" },
        @{ OS = "linux"; Arch = "amd64" },
        @{ OS = "linux"; Arch = "arm64" },
        @{ OS = "darwin"; Arch = "amd64" },
        @{ OS = "darwin"; Arch = "arm64" }
    )
} else {
    $targets = @(@{ OS = $Target; Arch = $Arch })
}

Push-Location $Backend
try {
    $ldflags = "-s -w"

    foreach ($t in $targets) {
        $os = $t.OS
        $arch = $t.Arch

        # 设置输出文件名
        if ($Output -and $targets.Count -eq 1) {
            $outputName = $Output
        } else {
            $outputName = Get-OutputName $os $arch
        }
        $OutputPath = Join-Path $ProjectRoot $outputName

        Write-Host "  构建 $os/$arch -> $outputName" -ForegroundColor Gray

        # 设置交叉编译环境变量
        $env:GOOS = $os
        $env:GOARCH = $arch
        $env:CGO_ENABLED = "0"

        go build -ldflags "$ldflags" -o $OutputPath .
        if ($LASTEXITCODE -ne 0) { throw "Go 构建失败: $os/$arch" }

        $FileSize = (Get-Item $OutputPath).Length / 1MB
        Write-Host "    文件大小: $([math]::Round($FileSize, 2)) MB" -ForegroundColor Gray
    }
} finally {
    # 恢复环境变量
    Remove-Item Env:GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
    Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue
    Pop-Location
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "  构建完成!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "使用示例:" -ForegroundColor Cyan
Write-Host "  .\build.ps1                    # 构建 Windows amd64" -ForegroundColor Gray
Write-Host "  .\build.ps1 -Target linux      # 构建 Linux amd64" -ForegroundColor Gray
Write-Host "  .\build.ps1 -Target linux -Arch arm64  # 构建 Linux arm64" -ForegroundColor Gray
Write-Host "  .\build.ps1 -Target all        # 构建所有平台" -ForegroundColor Gray
Write-Host "  .\build.ps1 -SkipFrontend      # 跳过前端构建" -ForegroundColor Gray
Write-Host ""
