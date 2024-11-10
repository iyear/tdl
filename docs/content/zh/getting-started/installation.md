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
iwr -useb https://docs.iyear.me/tdl/install.ps1 | iex
{{< /command >}}

#### 通过 `ghproxy.com` 镜像安装

{{< command >}}
$Script=iwr -useb https://docs.iyear.me/tdl/install.ps1;
$Block=[ScriptBlock]::Create($Script); Invoke-Command -ScriptBlock $Block -ArgumentList "", "$True"
{{< /command >}}

#### 安装特定版本

{{< command >}}
$Env:TDLVersion = "VERSION"
$Script=iwr -useb https://docs.iyear.me/tdl/install.ps1;
$Block=[ScriptBlock]::Create($Script); Invoke-Command -ScriptBlock $Block -ArgumentList "$Env:TDLVersion"
{{< /command >}}

{{< /tab >}}

{{< tab "macOS 和 Linux" >}}
`tdl` 将被安装到 `/usr/local/bin/tdl`，该脚本还可用于升级 `tdl`。

#### 安装最新版本

{{< command >}}
curl -sSL https://docs.iyear.me/tdl/install.sh | sudo bash
{{< /command >}}

#### 通过 `ghproxy.com` 镜像安装

{{< command >}}
curl -sSL https://docs.iyear.me/tdl/install.sh | sudo bash -s -- --proxy
{{< /command >}}

#### 安装特定版本

{{< command >}}
curl -sSL https://docs.iyear.me/tdl/install.sh | sudo bash -s -- --version VERSION
{{< /command >}}

{{< /tab >}}
{{< /tabs >}}

## 包管理器

{{< tabs "package managers" >}}

{{<tab "Homebrew" >}}
{{< command >}}
brew install telegram-downloader
{{< /command >}}
{{< /tab >}}

{{<tab "Scoop" >}}
{{< command >}}
scoop bucket add extras
scoop install telegram-downloader
{{< /command >}}
{{< /tab >}}

{{<tab "Termux" >}}
{{< command >}}
pkg install tdl
{{< /command >}}
{{< /tab >}}

{{<tab "AUR" >}}
{{< command >}}
yay -S tdl
{{< /command >}}
{{< /tab >}}

{{<tab "Nix" >}}

#### nix-env
{{< command >}}
nix-env -iA nixos.tdl
{{< /command >}}

#### NixOS-Configuration
```
environment.systemPackages = [
    pkgs.tdl
];
```

#### nix-shell
{{< command >}}
nix-shell -p tdl
{{< /command >}}

{{< /tab >}}

{{< /tabs >}}

[![Packaging status](https://repology.org/badge/vertical-allrepos/telegram-downloader.svg)](https://repology.org/project/telegram-downloader/versions)

## Docker

可用镜像：
- [`iyear/tdl`](https://hub.docker.com/r/iyear/tdl)
- [`ghcr.io/iyear/tdl`](https://ghcr.io/iyear/tdl)

可用标签：
- `latest`（默认）：最新的稳定版本
- `X.Y.Z`：`tdl`的特定版本

{{< tabs "docker" >}}
{{< tab "Docker" >}}

以一次性命令运行 `tdl`：
{{< command >}}
docker run --rm -it iyear/tdl <ARGUMENTS>
{{< /command >}}

进一步，挂载配置目录以保持持久化：
{{< command >}}
docker run --rm -it \
-v $HOME/.tdl:/root/.tdl \
iyear/tdl <ARGUMENTS>
{{< /command >}}

为了方便获取下载的文件，可以挂载下载目录和其他需要的目录：
{{< command >}}
docker run --rm -it \
-v $HOME/.tdl:/root/.tdl \
-v $HOME/Downloads:/downloads \
iyear/tdl <ARGUMENTS>
{{< /command >}}

在容器内运行 `tdl`：
{{< command >}}
docker run --rm -it <FLAGS> --entrypoint sh iyear/tdl
{{< /command >}}
{{< details title="预览输出" open=false >}}
```1
/ # tdl version
Version: 0.17.7
Commit: ace2402
Date: 2024-11-01T14:40:56+08:00

go1.21.13 linux/amd64
/ #
```
{{< /details >}}

如果希望使用 `localhost` 地址的代理，使用 `host` 网络运行：
{{< command >}}
docker run --rm -it <FLAGS> --network host iyear/tdl <ARGUMENTS>
{{< /command >}}
{{< /tab >}}

{{< tab "Docker Compose" >}}
使用 Docker Compose 运行 `tdl`，避免每次输入 `docker run` 选项。

{{< details title="docker-compose.yml" open=false >}}
{{< hint info >}}
示例配置使用 Docker Compose v2 语法。
{{< /hint >}}

```yaml
services:
  tdl:
    image: iyear/tdl # 或指定特定版本的 X.Y.Z 版本标签
    volumes:
      - $HOME/.tdl:/root/.tdl # 保持配置持久化
      - $HOME/Downloads:/downloads # 可选
      # - /path/to/your/need:/path/in/container
    stdin_open: true
    tty: true
    # 如果需要使用 localhost 地址的代理，使用 host 网络
    network_mode: host
```
{{< /details >}}

使用 Docker Compose 运行 `tdl`：
{{< command >}}
docker compose run --rm tdl <ARGUMENTS>
{{< /command >}}

在容器内运行 `tdl`：
{{< command >}}
docker compose run --rm --entrypoint sh tdl
{{< /command >}}
{{< details title="预览输出" open=false >}}
```1
/ # tdl version
Version: 0.17.7
Commit: ace2402
Date: 2024-11-01T14:40:56+08:00

go1.21.13 linux/amd64
/ #
```
{{< /details >}}

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

{{< tab "macOS" >}}
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
2. 安装 Go 的 1.21 版本或更高版本
3. 根据 Go 文档中的描述更新您的 `PATH` 环境变量

{{< hint info >}}
安装目录由 `GOPATH` 和 `GOBIN` 环境变量控制。如果设置了 `GOBIN`，则二进制文件将安装到该目录。如果设置了 `GOPATH`，则二进制文件将安装到 `GOPATH` 列表中第一个目录的 `bin` 子目录。否则，二进制文件将安装到默认的 `GOPATH` 的 `bin` 子目录（`$HOME/go` 或 `%USERPROFILE%\go`）。
{{< /hint >}}

然后构建：

{{< command >}}
go install github.com/iyear/tdl@latest
tdl version
{{< /command >}}
