## 🔍 Omarchy v1.13.1 - Doctor Safety Checks

### New Features

**🛡️ Safety Checks in `omarchy doctor`**

- Detects `.git` folders in home directory (prevents accidents)
- Warns about git repos on Desktop or Downloads
- Checks `node_modules` size (warns if >500MB)
- Shows disk space available
- Displays important environment variables
- Warns about large files tracked in git (>10MB)
- Counts VS Code extensions installed

**🔧 New Commands**

- `omarchy fix-home-git` - Remove accidental git repo from home

**🛡️ Git Sync Protection**

- Prevents `omarchy sync` from running in home directory
- Warns if .git exists in home

### Installation

```bash
go install github.com/Taha95-dev/Omarchy-CLI-tool@latest
```
