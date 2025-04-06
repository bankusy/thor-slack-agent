# Lightweight System Monitoring Agent

A lightweight system monitoring agent written in Golang that collects real-time system metrics and sends alerts to Slack when specific thresholds are exceeded.

---

### ✅ Features

- Real-time collection of:
  - CPU usage
  - Memory usage
  - Disk usage
  - Top 5 CPU-consuming processes
- Slack alert integration using Block Kit
- Configurable client ID
- Simple threshold-based alerting

## 📦 System Architecture

## 🚀 Getting Started

### 1. Create `.env` file

```env
CID=container-1 #Your Custom ID
MAX=10 #MAX_CPU_PERCENT
WEBHOOK_URL=https://hooks.slack.com/services/.../.../... #Webhook URL
```

### 2. Install dependencies
```go mod tidy```

### 3. Run the agent
```go run cmd/main.go```

## 🔔 Slack Alert Format
<img width="1014" alt="image" src="https://github.com/user-attachments/assets/8a967afc-1a15-4aa1-8035-6b77234e88d5" />

---
