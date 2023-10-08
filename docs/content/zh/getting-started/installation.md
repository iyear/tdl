---
title: "安装"
weight: 10
---

# 安装

## 一键脚本

{{< tabs "scripts" >}}

{{< tab "Windows" >}}
`tdl` 将被安装到 `$Env:SystemDrive\tdl`（将被添加到 `PATH` 中），该脚本还可用于升级 `tdl`。

#### 安装最新版本

{{< command >}}
iwr -useb https://ghproxy.com/https://raw.githubusercontent.com/iyear/tdl/master/scripts/install.ps1 | iex
{{< /command >}}

#### 通过 `ghproxy.com` 镜像安装

{{< command >}}
$Script=iwr -useb https://ghproxy.com/https://raw.githubusercontent.com/iyear/tdl/master/scripts/install.ps1;
$Block=[ScriptBlock]::Create($Script); Invoke-Command -ScriptBlock $Block -ArgumentList "", "$True"
{{< /command >}}

#### 安装特定版本

{{< command >}}
$Env:TDLVersion = "VERSION"
$Script=iwr -useb https://ghproxy.com/https://raw.githubusercontent.com/iyear/tdl/master/scripts/install.ps1;
$Block=[ScriptBlock]::Create($Script); Invoke-Command -ScriptBlock $Block -ArgumentList "$Env:TDLVersion"
{{< /command >}}

{{< /tab >}}

{{< tab "MacOS 和 Linux" >}}
`tdl` 将被安装到 `/usr/local/bin/tdl`，该脚本还可用于升级 `tdl`。

#### 安装最新版本

{{< command >}}
curl -sSL https://ghproxy.com/https://raw.githubusercontent.com/iyear/tdl/master/scripts/install.sh | sudo bash
{{< /command >}}

#### 通过 `ghproxy.com` 镜像安装

{{< command >}}
curl -sSL https://ghproxy.com/https://raw.githubusercontent.com/iyear/tdl/master/scripts/install.sh | sudo bash -s -- --proxy
{{< /command >}}

#### 安装特定版本

{{< command >}}
curl -sSL https://ghproxy.com/https://raw.githubusercontent.com/iyear/tdl/master/scripts/install.sh | sudo bash -s -- --version VERSION
{{< /command >}}

{{< /tab >}}
{{< /tabs >}}

## 包管理器

{{< tabs "package managers" >}}
{{< tab "Windows" >}}

#### Scoop

{{< command >}}
scoop bucket add extras
scoop install telegram-downloader
{{< /command >}}

{{< /tab >}}
{{< tab "MacOS" >}}
欢迎贡献！
{{< /tab >}}
{{< tab "Linux" >}}
欢迎贡献！
{{< /tab >}}
{{< /tabs >}}

## 预编译二进制

1. 下载指定操作系统和架构的压缩包：

{{< tabs "prebuilt" >}}
{{< tab "Windows" >}}
{{< button href="https://github.com/iyear/tdl/releases/latest/download/tdl_Windows_64bit.zip" >}}x86_64/amd64{{<
/button >}}
{{< button href="https://github.com/iyear/tdl/releases/latest/download/tdl_Windows_32bit.zip" >}}x86{{< /button >}}
{{< button href="https://github.com/iyear/tdl/releases/latest/download/tdl_Windows_arm64.zip" >}}arm64{{< /button >}}
{{< button href="https://github.com/iyear/tdl/releases/latest/download/tdl_Windows_armv5.zip" >}}armv5{{< /button >}}
{{< button href="https://github.com/iyear/tdl/releases/latest/download/tdl_Windows_armv6.zip" >}}armv6{{< /button >}}
{{< button href="https://github.com/iyear/tdl/releases/latest/download/tdl_Windows_armv7.zip" >}}armv7{{< /button >}}
{{< /tab >}}

{{< tab "MacOS" >}}
{{< button href="https://github.com/iyear/tdl/releases/latest/download/tdl_MacOS_64bit.tar.gz" >}}Intel{{< /button >}}
{{< button href="https://github.com/iyear/tdl/releases/latest/download/tdl_MacOS_arm64.tar.gz" >}}M1/M2{{< /button >}}
{{< /tab >}}

{{< tab "Linux" >}}
{{< button href="https://github.com/iyear/tdl/releases/latest/download/tdl_Linux_64bit.tar.gz" >}}x86_64/amd64{{<
/button >}}
{{< button href="https://github.com/iyear/tdl/releases/latest/download/tdl_Linux_32bit.tar.gz" >}}x86{{< /button >}}
{{< button href="https://github.com/iyear/tdl/releases/latest/download/tdl_Linux_arm64.tar.gz" >}}arm64{{< /button >}}
{{< button href="https://github.com/iyear/tdl/releases/latest/download/tdl_Linux_armv5.tar.gz" >}}armv5{{< /button >}}
{{< button href="https://github.com/iyear/tdl/releases/latest/download/tdl_Linux_armv6.tar.gz" >}}armv6{{< /button >}}
{{< button href="https://github.com/iyear/tdl/releases/latest/download/tdl_Linux_armv7.tar.gz" >}}armv7{{< /button >}}
{{< /tab >}}
{{< /tabs >}}

2. 解压缩压缩包
3. 将可执行文件移动到所需目录
4. 将此目录添加到 PATH 环境变量
5. 确保您对文件具有执行权限

## 源代码

要从源代码构建 `tdl` 的扩展版本，您必须：

1. 安装 [Git](https://git-scm.com/)
2. 安装 Go 的 1.19 版本或更高版本
3. 根据 Go 文档中的描述更新您的 `PATH` 环境变量

{{< hint info >}}
安装目录由 `GOPATH` 和 `GOBIN` 环境变量控制。如果设置了 `GOBIN`，则二进制文件将安装到该目录。如果设置了 `GOPATH`，则二进制文件将安装到 `GOPATH` 列表中第一个目录的 `bin` 子目录。否则，二进制文件将安装到默认的 `GOPATH` 的 `bin` 子目录（`$HOME/go` 或 `%USERPROFILE%\go`）。
{{< /hint >}}

然后构建：

{{< command >}}
go install github.com/iyear/tdl@latest
tdl version
{{< /command >}}
