# 🚀 Omarchy ━━━━ "Performance Update"

**One CLI to rule your dev workflow — git, scripts, env, cleanup, and more.**

[![Go Version](https://img.shields.io/badge/go-1.21%2B-blue)](https://go.dev/)
[![Clones](https://img.shields.io/badge/dynamic/json?color=brightgreen&label=clones&query=clones&url=https%3A%2F%2Fapi.github.com%2Frepos%2FTaha95-dev%2FOmarchy-CLI-tool%2Ftraffic%2Fclones)](https://github.com/Taha95-dev/Omarchy-CLI-tool)
[![Release](https://img.shields.io/github/v/release/Taha95-dev/Omarchy-CLI-tool)](https://github.com/Taha95-dev/Omarchy-CLI-tool/releases)

---

## ✨ Features

| Commands           | What Changed                                             |
| ------------------ | -------------------------------------------------------- |
| `omarchy doctor`   | Concurrent diagnostic suite for environment health.      |
| `omarchy clean-up` | Aggressive recursive purge for build artifacts and logs. |
| `omarchy run`      | Smart script runner (dev/build/test automation).         |
| `omarhcy sync`     | Secure Git orchestration with safety validation.         |
| `omarchy info`     | Instant project analytics (LOC, TODOs, Git health).      |
| `omarchy db`       | Database migrations and lifecycle management.            |

---

## 🚀 Quick Start

```bash
# Clone and build
git clone https://github.com/Taha95-dev/Omarchy-CLI-tool.git
cd Omarchy-CLI-tool
go build -o omarchy

# Run anywhere

./omarchy run dev
./omarchy info
./omarchy cleanup --docker

```

Or [download the latest release](https://github.com/Taha95-dev/Omarchy-CLI-tool/releases).

---

## 📖 Examples

### Run scripts without thinking

```bash
$ omarchy run
Available scripts:
  dev    → npm run dev
  build  → npm run build
  test   → npm test

$ omarchy run dev
🚀 Running: npm run dev
Server started on port 3000
```

### Get instant project insights

```bash
$ omarchy info
📊 Project Info
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📁 Files:         127
🧮 Lines:         8,452
🐛 TODOs:         3
🌿 Branch:        main
⏰ Last commit:   2 hours ago
```

### Git sync with safety

```bash
omarchy sync -a          # auto-commit + push
omarchy sync --tag v1.0  # commit + tag + push
```

---

## 🧠 Why Omarchy?

- **One tool** — no more switching between git, npm, docker, find, du, grep
- **Safety first** — won't let you commit your home folder or drop a production DB without confirmation
- **Fast** — written in Go, runs everywhere
- **Zero config** — works out of the box

---

## 🤝 Support

**Omarchy is built by a 13‑year‑old developer** — if you find it useful, consider giving it a ⭐ on Github.

---

## 📦 Installation

### From source

```bash
go install github.com/Taha95-dev/Omarchy-CLI-tool@latest
```

### From releases

Download the binary for your OS from [Releases](https://github.com/Taha95-dev/Omarchy-CLI-tool/releases).

---

## 📝 License

MIT — use it, learn from it, build something awesome.

---

## 🙌 Credits

Built by [Taha](https://github.com/Taha95-dev) — because building tools is better than waiting for them.

---
