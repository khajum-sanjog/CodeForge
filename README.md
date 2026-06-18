<div align="center">

# ⚡ CodeForge

### Next-Generation CI/CD Automation Platform

**Build • Deploy • Monitor • Rollback — All in One Engine**

<br/>

![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat-square)
![CI/CD](https://img.shields.io/badge/CI%2FCD-Automation-7C3AED?style=flat-square)
![Status](https://img.shields.io/badge/Status-Production-22C55E?style=flat-square)
![DevOps](https://img.shields.io/badge/DevOps-Platform-0EA5E9?style=flat-square)

</div>

---

## **🚀 The Problem**

Modern DevOps is fragmented.

Teams use:
- GitHub Actions
- Jenkins
- Docker scripts
- Manual SSH deploys
- Separate monitoring tools

👉 Result: Complexity, slow deployments, fragile systems.

---

## **⚡ The Solution — CodeForge**

CodeForge unifies your entire DevOps pipeline into a single automation engine.

> One tool. One language. One system.

---

## **🧠 KZM Deployment DSL**

```kzm
project "My API"

watch github "user/repo" on branch "main"

before deploy:
  run "npm install"
  run "npm test" must pass

deploy to ssh "ubuntu@server" at "/var/www/app":
  restart "pm2 restart app"
```
## 🔐 Secure Vault

Enterprise-grade secret management built-in.

AES-256-GCM encryption

Master key protection

Zero secret exposure in logs or UI

CLI-first management


```
codeforge secrets set AWS_KEY xxxx
codeforge secrets list


```
## 🔄 Rollback System

Never break production again.

Automatic snapshot before deploy

Instant rollback on failure

Versioned deployment recovery

## 📊 Real-Time Observability

Everything is visible in real time:

Live deployment logs

Pipeline status dashboard

Success / failure tracking

Full deployment history

## 🖥️ Unified Experience
CodeForge comes with everything built-in:

## ⚡ CLI for developers

🎨 Desktop GUI (Fyne)

## 🔄 Background daemon engine

### 🧭 System tray integration

### 🏗️ Architecture
```
CodeForge Engine
├── CLI (Cobra)
├── Daemon (Execution Engine)
├── GUI (Fyne Desktop)
├── KZM Parser (DSL Engine)
├── Executors (Deploy System)
├── Adapters (SSH / AWS / Docker / cPanel)
├── Secrets Vault (AES-256-GCM)
├── Logger (Streaming + JSON logs)
└── Rollback Engine
```

## 🌍 Supported Deploy Targets
SSH Servers,
AWS Lambda,
Docker Containers,
cPanel Hosting,
S3 Static Hosting,
VPS / Local Systems

## 📡 Example Pipeline
Code snippet
project "Node API"

watch github "user/api" on branch "main"

before deploy:
```
  run "npm install"
  run "npm test" must pass
```
```
deploy to ssh "ubuntu@server" at "/var/www/api":
  restart "pm2 restart api"
```
notify slack "#deployments"

## ⚡ Quick Start

# Install
```git clone [https://github.com/your-username/codeforge.git](https://github.com/your-username/codeforge.git)
cd codeforge

go mod tidy
go build -ldflags="-s -w" -o codeforge .
```
# Run
```
codeforge init
codeforge run my-api.kzm
codeforge daemon start
```
## 📂 Configuration
```
~/.codeforge/
pipelines/ → deployment configs
logs/      → execution logs
snapshots/ → rollback data
secrets.enc → encrypted vault
master.key → encryption key
```
## 🧠 Design Philosophy

Simplicity over complexity

Automation over manual work

Visibility over guesswork

Safety by default

Developer experience first


## 🛣️ Roadmap
Web dashboard (SaaS control panel)

Kubernetes integration

GitHub Actions importer

AI deployment assistant

Plugin marketplace

## 👨‍💻 Author
## KhajumSanjog


Built with ❤️ for developers who ship fast and safely.
