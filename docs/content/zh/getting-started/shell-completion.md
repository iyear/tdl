---
title: "自动补全"
weight: 30
---

# 自动补全

运行对应的命令以在所有会话中启用 Shell 自动补全：

{{< tabs "shell" >}}
{{< tab "bash" >}}

{{< command >}}
echo "source <(tdl completion bash)" >> ~/.bashrc
{{< /command >}}

{{< /tab >}}

{{< tab "zsh" >}}

{{< command >}}
echo "source <(tdl completion zsh)" >> ~/.zshrc
{{< /command >}}

{{< /tab >}}

{{< tab "fish" >}}

{{< command >}}
echo "tdl completion fish | source" >> ~/.config/fish/config.fish
{{< /command >}}

{{< /tab >}}

{{< tab "PowerShell" >}}

{{< command >}}
Add-Content -Path $PROFILE -Value "tdl completion powershell | Out-String | Invoke-Expression"
{{< /command >}}

{{< /tab >}}
{{< /tabs >}}
