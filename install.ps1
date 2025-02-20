$ErrorActionPreference = 'Stop'

$BinaryName = "mouse-keeper"
$Owner = "Hootrix"
$Repo = "go-mouse-keeper"

# 获取最新版本
$releases = "https://api.github.com/repos/$Owner/$Repo/releases"
$tag = (Invoke-WebRequest $releases | ConvertFrom-Json)[0].tag_name

# 检测系统架构
$arch = if ([Environment]::Is64BitOperatingSystem) { "x86_64" } else { "386" }
$fileName = "$BinaryName`_Windows_$arch.zip"
$downloadUrl = "https://github.com/$Owner/$Repo/releases/download/$tag/$fileName"

# 创建临时目录
$tempDir = Join-Path $env:TEMP ([System.Guid]::NewGuid())
New-Item -ItemType Directory -Path $tempDir | Out-Null

# 下载并解压
Write-Host "Downloading $BinaryName $tag for Windows $arch..."
Invoke-WebRequest $downloadUrl -OutFile "$tempDir\$fileName"
Expand-Archive "$tempDir\$fileName" -DestinationPath $tempDir

# 安装到用户目录
$installDir = "$env:USERPROFILE\.mouse-keeper"
$binary = "$installDir\$BinaryName.exe"

if (!(Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir | Out-Null
}

Move-Item "$tempDir\$BinaryName.exe" $binary -Force

# 添加到 PATH
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable(
        "Path",
        "$userPath;$installDir",
        "User"
    )
}

# 清理
Remove-Item $tempDir -Recurse -Force

Write-Host "$BinaryName has been installed successfully!"
Write-Host "You may need to restart your terminal for the PATH changes to take effect."
