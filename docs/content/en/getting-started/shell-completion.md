---
title: "Shell Completion"
weight: 30
---

# Shell Completion

Run corresponding command to enable shell completion in all sessions:

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
