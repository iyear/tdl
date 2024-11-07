---
title: "扩展 🆕"
weight: 70
---

# 扩展

{{< hint warning >}}
扩展是 tdl 的一项新功能，仍处于实验阶段，CLI 可能会在未来版本中发生变化。

如果你遇到任何问题或有任何建议，请在 GitHub 上[创建 Issue](https://github.com/iyear/tdl/issues/new/choose)。
{{< /hint >}}

## 概览

tdl 扩展是与 tdl 核心无缝集成的独立工具。它们提供了一种扩展 tdl 核心的方法，但不需要将每个新功能添加到核心代码中。

tdl 扩展具有以下特点：

- 它们可以添加和删除，而不会影响 tdl 核心。
- 它们与 tdl 集成，并会显示在 tdl 命令和其他地方。

tdl 扩展位于 `~/.tdl/extensions`，由 `tdl extension` 子命令控制。

使用以下命令快速体验 tdl 扩展：

{{< command >}}
tdl extension install iyear/tdl-whoami
{{< /command >}}

{{< command >}}
tdl whoami
{{< /command >}}

你可以看到 `tdl-whoami` 扩展的输出。详情请参阅 [tdl-whoami](https://github.com/iyear/tdl-whoami)。
```
You are XXXXX. ID: XXXXXXXX
```

## 查找扩展

你可以通过浏览[带有 `tdl-extension` 主题的代码库](https://github.com/topics/tdl-extension)来查找扩展。

## 安装扩展

要安装扩展，请使用 `extension install` 子命令。

扩展有两种类型：

- `GitHub` : 托管在 GitHub 代码库上的扩展。

    {{< command >}}
    tdl extension install <owner>/<repo>
    {{< /command >}}

    要从私有代码库安装扩展，必须设置 `GITHUB_TOKEN` 环境变量为 [GitHub 个人访问令牌](https://github.com/settings/personal-access-tokens/new)（具有 `Contents` 读取权限）。

    {{< command >}}
    export GITHUB_TOKEN=YOUR_TOKEN
    tdl extension install <owner>/<private-repo>
    {{< /command >}}

- `Local` : 存储在本地计算机上的扩展。

    {{< command >}}
    tdl extension install /path/to/extension
    {{< /command >}}

强制安装已经存在的扩展，请使用 `--force` 选项：

{{< command >}}
tdl extension install --force EXTENSION
{{< /command >}}

一次安装多个扩展，请使用以下命令：

{{< command >}}
tdl extension install <owner>/<repo1> /path/to/extension2 ...
{{< /command >}}

仅打印信息而不实际安装扩展，请使用 `--dry-run` 选项：

{{< command >}}
tdl extension install --dry-run EXTENSION
{{< /command >}}

如果你已经安装了同名的扩展，安装将失败。例如，如果你已经安装了 `foo/tdl-whoami`，则必须在安装 `bar/tdl-whoami` 之前卸载它。

## 运行扩展

安装扩展后，可以像运行本地 tdl 命令一样运行扩展，使用 `tdl EXTENSION-NAME`。`EXTENSION-NAME` 是包含扩展的代码库的名称，去掉 `tdl-` 前缀。

例如，如果你从 `iyear/tdl-whoami` 代码库安装了扩展，可以使用以下命令运行扩展。

{{< command >}}
tdl whoami
{{< /command >}}

运行扩展时，全局配置仍然可用。例如，以下命令在运行 `tdl-whoami` 扩展时指定命名空间和代理。

{{< command >}}
tdl -n foo --proxy socks5://localhost:1080 whoami
{{< /command >}}

扩展自身的选项也可以使用。例如，以下命令在运行 `tdl-whoami` 扩展时启用详细模式。

{{< hint info >}}
请记住在扩展子命令之前写全局选项，在扩展子命令之后写扩展选项：
{{< command >}}
tdl <全局选项> <扩展名> <扩展选项>
{{< /command >}}

{{< /hint >}}

{{< command >}}
tdl -n foo whoami -v
{{< /command >}}

通常可以在包含扩展的代码库的 README 中找到有关如何使用扩展的具体信息。

## 查看已安装的扩展

要查看所有已安装的扩展，请使用 `extension list` 子命令。此命令将列出所有已安装的扩展及其作者和版本。

{{< command >}}
tdl extension list
{{< /command >}}

## 更新扩展

要更新扩展，请使用 `extension upgrade` 子命令。将 `EXTENSION` 参数替换为扩展的名称。

{{< command >}}
tdl extension upgrade EXTENSION1 EXTENSION2 ...
{{< /command >}}

更新所有已安装的扩展，请设置 `EXTENSION` 参数为空。

{{< command >}}
tdl extension upgrade
{{< /command >}}

从 GitHub 私有代码库升级扩展，必须设置 `GITHUB_TOKEN` 环境变量为 [GitHub 个人访问令牌](https://github.com/settings/personal-access-tokens/new)（具有 `Contents` 读取权限）。

{{< command >}}
export GITHUB_TOKEN=YOUR_TOKEN
tdl extension upgrade EXTENSION
{{< /command >}}

仅打印信息而不实际升级扩展，请使用 `--dry-run` 选项：

{{< command >}}
tdl extension upgrade --dry-run EXTENSION
{{< /command >}}

## 卸载扩展

要卸载扩展，请使用 `extension remove` 子命令。将 `EXTENSION` 参数替换为扩展的名称。

{{< command >}}
tdl extension remove EXTENSION1 EXTENSION2 ...
{{< /command >}}

仅打印信息而不实际卸载扩展，请使用 `--dry-run` 选项：

{{< command >}}
tdl extension remove --dry-run EXTENSION
{{< /command >}}

## 开发扩展

请参阅 [tdl-extension-template](https://github.com/iyear/tdl-extension-template) 代码库，了解如何为 tdl 创建、构建和发布扩展。
