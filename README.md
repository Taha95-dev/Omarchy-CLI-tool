````markdown
[![Go Reference](https://img.shields.io/badge/go-reference-blue.svg)](https://pkg.go.dev/github.com/Taha95-dev/Omarchy-CLI-tool)
[![Release](https://img.shields.io/github/v/release/Taha95-dev/Omarchy-CLI-tool)](https://github.com/Taha95-dev/Omarchy-CLI-tool/releases)
[![Tests](https://github.com/Taha95-dev/Omarchy-CLI-tool/actions/workflows/test.yml/badge.svg)](https://github.com/Taha95-dev/Omarchy-CLI-tool/actions)

# 🚀 Omarchy

**One command to scaffold full-stack projects.**

Omarchy is a CLI tool that creates complete project structures with React frontends, Node/Go backends, Git initialization, and more.

## ✨ Features

- 📁 **Project templates** - web, cli, fullstack, backend
- ⚛️ **React frontend** with Vite
- 🐹 **Go backend** with API starter
- 📦 **Node backend** with Express
- 🔧 **Git initialization** with .gitignore
- 📝 **Config file** (~/.omarchy.yaml)
- 🔄 **Git sync** - one-command commit + push
- 🩺 **Doctor** - check your development environment
- 📊 **File counter** - count files by extension
- 🌳 **Tree view** - visualize folder structure
- 🖥️ **Cross-platform** - Windows, Linux, macOS

## 📦 Installation

### Windows

```bash
# Download from Releases
# Or build from source
go install github.com/Taha95-dev/Omarchy-CLI-tool@latest
```
````

## 🚀 Quick Start Examples

```bash
# Create a React + Node fullstack app
omarchy -name my-app -type fullstack -react -node -git

# Create a Vue + Go fullstack app
omarchy -name my-app -type fullstack -vue -go

# Create a Next.js app
omarchy -name my-app -type fullstack -next -node

# Create a backend-only API
omarchy -name my-api -type backend -node
```
# Test 2FA setup
