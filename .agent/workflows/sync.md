---
description: Sync code to GitHub with a meaningful, AI-generated commit message
---

1. Check status and stage changes.
// turbo
```powershell
git status
git add .
```

2. **Generate a meaningful commit message** based on the `git status` output (e.g., `feat: details` or `fix: details`).
   - Run `git commit -m "your_message"` using `SafeToAutoRun=true` to comply with the "no approval" rule.

3. Push changes.
// turbo
```powershell
git push
```
