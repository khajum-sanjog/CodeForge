<div align="center">

# ⚡ CodeForge

<img src="./logo.png" width="120" />

### CI/CD Automation for Modern Developers

**Build • Test • Deploy • Monitor — in one unified DevOps engine**

<br/>

![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat-square)
![CI/CD](https://img.shields.io/badge/CI%2FCD-Automation-7C3AED?style=flat-square)
![Status](https://img.shields.io/badge/Status-Active-22C55E?style=flat-square)

---

</div>

## ✨ What is CodeForge?

**CodeForge is a next-generation DevOps automation platform** designed to remove complexity from software delivery.

It turns your repositories into **self-deploying systems** with minimal configuration and maximum control.

> Think: **GitHub Actions + Jenkins + Vercel + custom deployment engine — reimagined for developers who ship fast.**

---

## 🚀 Why CodeForge?

Modern DevOps is fragmented.

CodeForge unifies everything into one system:

- 🧩 Build pipelines
- 🔁 Automated deployments
- 🔐 Secure secrets management
- 📦 Rollback system
- 📡 Real-time logs & monitoring
- 🧠 Human-readable pipeline language (KZM)
- 🖥️ CLI + GUI + Daemon in one binary

---

## ⚙️ Core Features

### ⚡ CI/CD Engine
- GitHub / GitLab integration
- Folder watchers
- Cron-based scheduling
- Manual & API triggers

---

### 🧠 KZM Pipeline Language

A simple, human-readable deployment DSL:

```kzm
project "My API"

watch github "user/repo" on branch "main"

before deploy:
  run "npm install"
  run "npm test" must pass

deploy to ssh "ubuntu@server" at "/var/www/app":
  restart "pm2 restart app"
🔐 Secure Vault

Encrypted secrets using AES-256-GCM

Master key protection
CLI-based management
Zero secret exposure in logs or UI
codeforge secrets set AWS_KEY xxxx
codeforge secrets list
🔄 Rollback System
Snapshot before every deploy
Instant rollback on failure
Safe recovery at any point
📊 Real-Time Observability
Live deployment logs
Pipeline status dashboard
Success / failure tracking
Full deployment history
🖥️ Unified Experience
⚡ CLI for developers
🎨 Beautiful GUI (Fyne)
🔄 Background daemon
🧭 System tray integration
🏗️ Architecture
CodeForge
├── CLI (Cobra)
├── Daemon (Pipeline Engine)
├── GUI (Fyne Desktop App)
├── KZM Parser (DSL Engine)
├── Executors (Build & Deploy)
├── Adapters (SSH, AWS, Docker, cPanel)
├── Secrets Vault (AES-256-GCM)
├── Logger (Streaming + JSON logs)
└── Rollback Engine
⚙️ Installation
🔧 Requirements
Go 1.22+
Git
Linux / macOS / Windows
📦 Install from source
git clone https://github.com/your-username/codeforge.git
cd codeforge

go mod tidy
go build -ldflags="-s -w" -o codeforge .
🚀 Install globally
sudo mv codeforge /usr/local/bin/
codeforge --version
🧪 Quick Start
codeforge init
codeforge check my-api.kzm
codeforge run my-api.kzm
codeforge daemon start
📂 Configuration

All system data lives in:

~/.codeforge/
Structure
pipelines/   → deployment definitions (.kzm)
logs/        → execution logs
snapshots/   → rollback data
secrets.enc  → encrypted secrets vault
master.key   → encryption key
daemon.pid   → running process
🔐 Secrets
codeforge secrets set GITHUB_TOKEN ghp_xxx
codeforge secrets set SLACK_WEBHOOK https://hooks.slack.com/...
codeforge secrets list

🔒 Secrets are always encrypted and never exposed anywhere.

🌍 Supported Deploy Targets
SSH Servers
AWS Lambda
Docker Containers
cPanel Hosting
S3 Static Hosting
VPS / Local Systems
📡 Example Pipeline
project "Node API"

watch github "user/api" on branch "main"

before deploy:
  run "npm install"
  run "npm test" must pass

deploy to ssh "ubuntu@server" at "/var/www/api":
  restart "pm2 restart api"

notify slack "#deployments"
🖥️ CLI Commands
codeforge gui
codeforge init
codeforge check file.kzm
codeforge run file.kzm
codeforge daemon start
codeforge status
codeforge logs my-api --tail
codeforge trigger my-api
codeforge rollback my-api
🔔 Notifications
Slack integration
Email alerts
Deployment reports
Failure notifications
🧠 Design Philosophy

We believe:

Simplicity > Complexity
Automation > Manual work
Visibility > Guesswork
Safety first
Developer experience matters
🛣️ Roadmap
 Web dashboard (SaaS version)
 Kubernetes adapter
 GitHub Actions import
 AI deployment assistant
 Multi-region deployments
 Plugin marketplace
👨‍💻 Author
KhajumSanjog

Built with ❤️ for developers who ship fast and break nothing.
