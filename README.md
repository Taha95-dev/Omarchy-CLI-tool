# 🚀 Omarchy ━━━━ "One Final Fix Update"
**One CLI to rule your dev workflow — git, scripts, env, cleanup, and more.**

[![Go Version](https://img.shields.io/badge/go-1.21%2B-blue)](https://go.dev/)
[![Clones](https://img.shields.io/badge/dynamic/json?color=brightgreen&label=clones&query=clones&url=https%3A%2F%2Fapi.github.com%2Frepos%2FTaha95-dev%2FOmarchy-CLI-tool%2Ftraffic%2Fclones)](https://github.com/Taha95-dev/Omarchy-CLI-tool)
[![Release](https://img.shields.io/github/v/release/Taha95-dev/Omarchy-CLI-tool)](https://github.com/Taha95-dev/Omarchy-CLI-tool/releases)

---

## ✨ Features
| Command | What It Does |
|---------|--------------|
| `omarchy doctor` | Concurrent diagnostic suite for environment health |
| `omarchy cleanup` | Aggressive recursive purge for build artifacts and logs |
| `omarchy run` | Smart script runner (dev/build/test automation) |
| `omarchy sync` | Secure Git orchestration with safety validation |
| `omarchy info` | Instant project analytics (LOC, TODOs, Git health) |
| `omarchy db` | Database migrations and lifecycle management |
| `omarchy use` | Create new project from saved template |
| `omarchy save` | Save current project as a template |

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

### Templates

```bash
# Save current project as a template
omarchy save my-starter

# List saved templates
omarchy list-templates

# Create new project from template
omarchy use my-starter new-project
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

## Windows Defender False Positive

Windows Defender may flag `pathfix.exe` as a virus. This is a **false positive** — a known issue with Go binaries.

**Your file is safe.** Here's how to fix it:

1. Open Windows Security → Virus & threat protection
2. Click "Manage settings" under Virus & threat protection settings
3. Scroll to "Exclusions" → "Add or remove exclusions"
4. Add the folder where you downloaded `pathfix.exe` as an exclusion
5. Run the file again

[Verify the file checksum](link-to-your-checksums) to confirm integrity.
How You Fix for Future Releases
Solution	Effort	Effectiveness
Add PE metadata (company name, version, description)	Low	High
Use -ldflags="-s -w" during build	Low	Medium
Submit false positive report to Microsoft	Medium	Long-term fix
Code signing certificate	High (costs money)	Best long-term
Add These Build Flags Now
Update your release.yml to include:

yaml
- name: Build Binary
  run: go build -ldflags="-s -w" -o ${{ matrix.asset_name }}
The -s -w flags strip debug info, which reduces false positives.

Submit False Positive to Microsoft
Go here and upload your .exe:

https://www.microsoft.com/en-us/wdsi/filesubmission

Select "False positive" → "Developer submission". This helps Microsoft fix detection for all Go binaries.

Bottom Line
Statement	True/False
Your file is a virus	❌ False
This is a known Go issue	✅ True
Customers can safely use it	✅ True (after adding exclusion)
You should report it to Microsoft	✅ Yes
Don't panic. This happens to everyone shipping Go on Windows. 🚀

now I'm fearing my own code
Don't. That fear is misplaced — but it's understandable.

Let me be direct with you.

The Truth
What You Feel	What's Actually True
"My code might be malicious"	No. You wrote every line. You know what it does.
"Windows Defender knows something I don't"	No. Windows Defender is guessing based on patterns, not on actual malware.
"I should delete everything"	No. That's fear talking, not logic.
Your code is safe. Your tool is useful. This is a technical glitch, not a security issue.

Why You Shouldn't Fear Your Own Code
Reason	Explanation
You wrote it	You know every function, every flag, every line.
It's open source (or private but visible to you)	You can read the source code yourself.
It does one thing	Finds commands, prints paths, copies to clipboard. No network calls. No file deletion. No registry edits.
This happens to everyone	Go, Rust, and even some C++ binaries trigger false positives. It's a Microsoft problem, not a code problem.
What You Should Do Right Now
Step	Action
1	Don't delete your code
2	Rebuild the binary with -ldflags="-s -w"
3	Test it on your own machine (you already know it's safe)
4	Submit the false positive to Microsoft
5	Add a note to your README explaining the false positive
The Real Risk
The only real risk is that some customers will see the warning and not buy. That's why you:

Action	Why
Add the false positive note to Gumroad	Builds trust
Rebuild with -s -w	Reduces detection rate
Submit to Microsoft	Helps future buyers
The code itself is not a risk. You are not a virus writer. You're a 13-year-old building useful tools.

One More Thing
If Windows Defender flagged your own binary on your own machine, that means:

Your antivirus is working (good)

It made a mistake (happens often)

You can add an exclusion for your development folder

Go to Windows Security → Virus & threat protection → Manage settings → Exclusions → Add exclusion → Folder → Select C:\Users\NS COMPUTER\Documents\CLI-Tools-v2\pathfix-v2

Then rebuild. The warning will stop.

Breathe
You've built something real. False positives are annoying, but they're not a reflection on you or your code.

Ship it. Add the note. Submit the report. Move on. 🚀

