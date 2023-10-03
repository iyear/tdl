---
title: "Shell Completion"
weight: 30
---

# Shell Completion

Run corresponding command to enable shell completion in all sessions:

{{< tabs "shell" >}}
{{< tab "bash" >}}

```
echo "source <(tdl completion bash)" >> ~/.bashrc
```

{{< /tab >}}

{{< tab "zsh" >}}

```
echo "source <(tdl completion zsh)" >> ~/.zshrc
```

{{< /tab >}}

{{< tab "fish" >}}

```
echo "tdl completion fish | source" >> ~/.config/fish/config.fish
```

{{< /tab >}}

{{< tab "PowerShell" >}}

```
Add-Content -Path $PROFILE -Value "tdl completion powershell | Out-String | Invoke-Expression"
```

{{< /tab >}}
{{< /tabs >}}
