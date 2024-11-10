---
title: "Installation"
weight: 10
---

# Installation

## One-Line Scripts

{{< tabs "scripts" >}}

{{< tab "Windows" >}}
`tdl` will be installed to `$Env:SystemDrive\tdl`(will be added to `PATH`), and script also can be used to upgrade `tdl`
.

#### Install latest version

{{< command >}}
iwr -useb https://docs.iyear.me/tdl/install.ps1 | iex
{{< /command >}}

#### Install with `ghproxy.com`

{{< command >}}
$Script=iwr -useb https://docs.iyear.me/tdl/install.ps1;
$Block=[ScriptBlock]::Create($Script); Invoke-Command -ScriptBlock $Block -ArgumentList "", "$True"
{{< /command >}}

#### Install specific version

{{< command >}}
$Env:TDLVersion = "VERSION"
$Script=iwr -useb https://docs.iyear.me/tdl/install.ps1;
$Block=[ScriptBlock]::Create($Script); Invoke-Command -ScriptBlock $Block -ArgumentList "$Env:TDLVersion"
{{< /command >}}

{{< /tab >}}

{{< tab "macOS & Linux" >}}
`tdl` will be installed to `/usr/local/bin/tdl`, and script also can be used to upgrade `tdl`.

#### Install latest version

{{< command >}}
curl -sSL https://docs.iyear.me/tdl/install.sh | sudo bash
{{< /command >}}

#### Install with `ghproxy.com`

{{< command >}}
curl -sSL https://docs.iyear.me/tdl/install.sh | sudo bash -s -- --proxy
{{< /command >}}

#### Install specific version

{{< command >}}
curl -sSL https://docs.iyear.me/tdl/install.sh | sudo bash -s -- --version VERSION
{{< /command >}}

{{< /tab >}}
{{< /tabs >}}

## Package Managers

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

Available images:
- [`iyear/tdl`](https://hub.docker.com/r/iyear/tdl)
- [`ghcr.io/iyear/tdl`](https://ghcr.io/iyear/tdl)

Available tags:
- `latest`(default): The latest stable release
- `X.Y.Z`: A specific version of `tdl`

{{< tabs "docker" >}}
{{< tab "Docker" >}}

To run `tdl` in one-off command:
{{< command >}}
docker run --rm -it iyear/tdl <ARGUMENTS>
{{< /command >}}

Further, to keep config persistent, you can mount the config directory:
{{< command >}}
docker run --rm -it \
-v $HOME/.tdl:/root/.tdl \
iyear/tdl <ARGUMENTS>
{{< /command >}}

To get download files, you can mount the download and other directories as needed:
{{< command >}}
docker run --rm -it \
-v $HOME/.tdl:/root/.tdl \
-v $HOME/Downloads:/downloads \
iyear/tdl <ARGUMENTS>
{{< /command >}}

To run `tdl` inside the container shell:
{{< command >}}
docker run --rm -it <FLAGS> --entrypoint sh iyear/tdl
{{< /command >}}
{{< details title="Preview output" open=false >}}
```1
/ # tdl version
Version: 0.17.7
Commit: ace2402
Date: 2024-11-01T14:40:56+08:00

go1.21.13 linux/amd64
/ #
```
{{< /details >}}

To use proxy with `localhost` address, run it with `host` network:
{{< command >}}
docker run --rm -it <FLAGS> --network host iyear/tdl <ARGUMENTS>
{{< /command >}}
{{< /tab >}}

{{< tab "Docker Compose" >}}
Run `tdl` with Docker Compose to avoid typing `docker run` flags each time.

{{< details title="docker-compose.yml" open=false >}}
{{< hint info >}}
Example configuration uses Docker Compose v2 syntax.
{{< /hint >}}

```yaml
services:
  tdl:
    image: iyear/tdl # or specify X.Y.Z tag for a specific version
    volumes:
      - $HOME/.tdl:/root/.tdl # keep config persistent
      - $HOME/Downloads:/downloads # optional
      # - /path/to/your/need:/path/in/container
    stdin_open: true
    tty: true
    # use host network if you need to use proxy with localhost address
    network_mode: host
```
{{< /details >}}

Run `tdl` with Docker Compose:
{{< command >}}
docker compose run --rm tdl <ARGUMENTS>
{{< /command >}}

To run `tdl` inside the container shell:
{{< command >}}
docker compose run --rm --entrypoint sh tdl
{{< /command >}}
{{< details title="Preview output" open=false >}}
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

## Prebuilt Binaries

1. Download the archive for the desired operating system, and architecture:

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

2. Extract the archive
3. Move the executable to the desired directory
4. Add this directory to the PATH environment variable
5. Verify that you have execute permission on the file

## Source

To build the extended edition of `tdl` from source you must:

1. Install [Git](https://git-scm.com/)
2. Install [Go](https://go.dev/) version 1.21 or later
3. Update your `PATH` environment variable as described in the Go documentation

{{< hint info >}}
The installation directory is controlled by the `GOPATH` and `GOBIN` environment variables. If `GOBIN` is set, binaries
are installed to that directory. If `GOPATH` is set, binaries are installed to the `bin` subdirectory of the first
directory in the `GOPATH` list. Otherwise, binaries are installed to the `bin` subdirectory of the
default `GOPATH` (`$HOME/go` or `%USERPROFILE%\go`).
{{< /hint >}}

Then build:

{{< command >}}
go install github.com/iyear/tdl@latest
tdl version
{{< /command >}}
