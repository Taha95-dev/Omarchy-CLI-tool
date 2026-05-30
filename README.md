## 🎉 Omarchy v2.3.0 - No More Tears Update

### 🛡️ Database Safety (Finally)

| Command                     | What changed                             |
| --------------------------- | ---------------------------------------- |
| `omarchy db reset`          | Auto dry-run → "DELETE" → "I UNDERSTAND" |
| `omarchy db migrate` (prod) | Requires `--dry-run` + `--force`         |
| `omarchy db seed` (prod)    | "SEED" confirmation required             |

**No more accidental production database wipes.**

### ⚡ Git Improvements

| Command                     | What it does                                   |
| --------------------------- | ---------------------------------------------- |
| `omarchy sync --tag v2.3.0` | Commit + tag + push in one command             |
| `omarchy branch -d <name>`  | Safe branch deletion (checks merged, confirms) |
| `omarchy branch -D <name>`  | Force delete unmerged branches                 |

### 📦 Install

```bash
go install github.com/Taha95-dev/Omarchy-CLI-tool@latest
```

## 💬 Feedback & Support

I'd love to hear about your experience with Omarchy!

** Your feedback helps me improve: **

- What worked well?
- What was confusing?
- What feature would you like to see?

### Ways to share your feedback:

| Method                 | Where                                                                                |
| ---------------------- | ------------------------------------------------------------------------------------ |
| **GitHub Issues**      | [Open an issue](https://github.com/Taha95-dev/Omarchy-CLI-tool/issues/new)           |
| **GitHub Discussions** | [Start a discussion](https://github.com/Taha95-dev/Omarchy-CLI-tool/discussions/new) |
| **Email**              | [kashiftaha976@gmail.com](mailto:kashiftaha976@gmail.com)                            |

### Star the project

If Omarchy helps you, consider giving it a ⭐ on GitHub!

[![GitHub stars](https://img.shields.io/github/stars/Taha95-dev/Omarchy-CLI-tool?style=social)](https://github.com/Taha95-dev/Omarchy-CLI-tool)
