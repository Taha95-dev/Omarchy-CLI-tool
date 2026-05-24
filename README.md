## 🚀 Omarchy v2.0.0 - The Major Release

### What's New

**✨ Deploy Command**

- One-command deployment to Render, Netlify, Vercel
- Auto-detects project type
- Creates platform configs

**🗄️ Database Management**

- `omarchy db init` - Initialize database
- `omarchy db migrate` - Run migrations
- `omarchy db migrate --dry-run` - Preview safely
- `omarchy db seed` - Seed data
- `omarchy db reset` - Reset (with --dry-run)
- Supports PostgreSQL, MySQL, SQLite

**🐍 Python Backend**

- FastAPI framework
- Docker support
- Requirements.txt included

**⚙️ C++ Backend**

- Crow web framework
- CMake build system
- Docker support

**🌳 Tree Improvements**

- `omarchy tree --depth N` - Limit depth
- Skips hidden folders and node_modules

**🛡️ Safety Features**

- `--dry-run` for destructive operations
- Better error messages
- Confirmation prompts

### Installation

```bash
go install github.com/Taha95-dev/Omarchy-CLI-tool@latest
```
