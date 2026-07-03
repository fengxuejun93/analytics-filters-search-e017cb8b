# 二手货品置换平台 - 启动脚本
# 使用方式: .\start.ps1
# 访问地址: http://localhost:8080

$ErrorActionPreference = "Stop"

Write-Host "=== 二手货品置换平台 ===" -ForegroundColor Cyan
Write-Host ""

# 编译
Write-Host "[1/2] 编译项目..." -ForegroundColor Yellow
go build -o server.exe .
if ($LASTEXITCODE -ne 0) {
    Write-Host "编译失败！" -ForegroundColor Red
    exit 1
}
Write-Host "编译成功" -ForegroundColor Green

# 启动
Write-Host "[2/2] 启动服务..." -ForegroundColor Yellow
Write-Host ""
Write-Host "访问地址: http://localhost:8080" -ForegroundColor Cyan
Write-Host "按 Ctrl+C 停止服务" -ForegroundColor Gray
Write-Host ""

.\server.exe
