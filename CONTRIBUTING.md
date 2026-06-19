# Contributing to Omarchy-CLI-Tool

First off, thank you for considering contributing to Omarchy! 🎉 It means a lot that you want to help make this tool better.

## Table of Contents
- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Reporting Bugs](#reporting-bugs)
- [Suggesting Features](#suggesting-features)
- [Pull Requests](#pull-requests)
- [Development Setup](#development-setup)
- [Style Guide](#style-guide)
- [Testing](#testing)
- [License](#license)

---

## Code of Conduct

This project and everyone participating in it is governed by the [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

---

## How Can I Contribute?

### 🐛 Reporting Bugs
If you find a bug, please create an issue with:
- A clear, descriptive title
- Steps to reproduce the bug
- What you expected to happen
- What actually happened
- Your OS and version
- Any error messages or logs

### 💡 Suggesting Features
Have an idea? Open an issue with:
- A clear, descriptive title
- A detailed explanation of the feature
- Why it would be useful
- Any examples or mockups

### 🔧 Pull Requests
1. Fork the repository
2. Create a new branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`go test ./...`)
5. Commit your changes (`git commit -m 'Add some amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

---

## Development Setup

### Prerequisites
- Go 1.21 or higher
- Git

### Clone and Build
```bash
git clone https://github.com/Taha95-dev/Omarchy-CLI-tool.git
cd Omarchy-CLI-tool
go build -o omarchy
Run Tests
bash
go test ./...
Run with Coverage
bash
go test -cover ./...
Style Guide
Go Code
Follow standard Go conventions

Run go fmt before committing

Use meaningful variable names

Add comments for exported functions

Commit Messages
Use the present tense ("Add feature" not "Added feature")

Use the imperative mood ("Move cursor to..." not "Moves cursor to...")

Limit the first line to 72 characters

Reference issues and pull requests liberally after the first line

Example:
text
Add support for custom templates

- Add `omarchy save` command
- Add `omarchy use` command
- Update README with examples
- Fix #42
Testing
Running Tests
bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package
go test ./pkg/templates
Writing Tests
Write tests for new features

Update tests when changing behavior

Use table-driven tests where appropriate

License
By contributing, you agree that your contributions will be licensed under the MIT License.

Questions?
If you have any questions, feel free to:

Open an issue

Reach out to Taha95-dev

Thank you for contributing! 🙌
