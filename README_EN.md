# X-Bot: Hackathon Tweet Auto-Reply System

A Go-based automated Twitter bot that monitors tweets from followed users, identifies hackathon-related content using LLM, and automatically replies with pre-configured promotional messages.

## ✨ Features

- **Auto Monitoring**: Fetch latest tweets from followed users
- **Smart Detection**: Use LLM (GPT) to identify hackathon-related tweets
- **Auto Reply**: Automatically reply with preset ad copy on relevant tweets
- **Ad Management**: Support multiple ad copies with priority-based rotation
- **Scheduled Tasks**: Cron-based workflow scheduling
- **Analytics**: Reply logging and statistics
- **REST API**: Manual trigger and management via RESTful API

## 🏗️ Architecture

```
├── cmd/server/          # Application entry point
├── config/              # Configuration files
├── migrations/          # Database migration scripts
├── internal/
│   ├── config/          # Configuration loading
│   ├── domain/          # Domain layer (entities + repository interfaces)
│   ├── infrastructure/  # Infrastructure layer (DB/Twitter/LLM clients)
│   ├── application/     # Application layer (business logic)
│   └── interfaces/      # Interface layer (HTTP API + scheduler)
└── pkg/                 # Shared utilities
```

**Tech Stack:**
- Go 1.21+
- PostgreSQL
- Gin (HTTP framework)
- GORM (ORM)
- Twitter API v2 (OAuth 1.0a)
- OpenAI API (LLM)

## 🚀 Quick Start

### 1. Prerequisites

- Go 1.21+
- PostgreSQL 12+
- Twitter Developer Account (API v2 access)
- OpenAI API Key

### 2. Clone Repository

```bash
git clone https://github.com/zhoubofsy/x-bot.git
cd x-bot
```

### 3. Set Environment Variables

```bash
# Database
export DB_PASSWORD=your_db_password

# Twitter API (OAuth 1.0a)
export TWITTER_API_KEY=your_api_key
export TWITTER_API_SECRET=your_api_secret
export TWITTER_ACCESS_TOKEN=your_access_token
export TWITTER_ACCESS_SECRET=your_access_secret
export TWITTER_BEARER_TOKEN=your_bearer_token

# OpenAI
export OPENAI_API_KEY=your_openai_api_key

# API Authentication (optional)
export API_KEY=your_api_key_for_http_auth
```

### 4. Initialize Database

```bash
# Create database
createdb x_bot

# Run migration script
psql -U postgres -d x_bot -f migrations/001_init.sql
```

### 5. Build and Run

```bash
# Install dependencies
go mod tidy

# Build
go build -o bin/x-bot ./cmd/server/

# Run
./bin/x-bot -config config/config.yaml
```

## 📝 Configuration

Edit `config/config.yaml`:

```yaml
server:
  port: 8080
  mode: release  # debug | release

database:
  host: localhost
  port: 5432
  user: postgres
  password: ${DB_PASSWORD}
  dbname: x_bot

workflow:
  default_tweet_count: 10      # Tweets per user
  reply_interval: 60s          # Interval between replies
  max_daily_replies: 100       # Max replies per day
  enable_scheduler: true       # Enable scheduled tasks
  schedule: "0 */2 * * *"      # Cron expression (every 2 hours)
```

## 🔌 API Reference

All API requests require authentication header:
```
Authorization: Bearer <API_KEY>
```

### Workflow

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/workflow/execute` | Execute workflow |
| POST | `/api/v1/workflow/sync-following` | Sync following list |

**Execute Workflow Parameters:**
```json
{
  "tweet_count": 10,
  "dry_run": false
}
```

### Statistics & Logs

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/stats` | Get statistics |
| GET | `/api/v1/reply-logs?limit=20` | Get reply logs |

### Ad Copy Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/ad-copies` | List all ad copies |
| GET | `/api/v1/ad-copies/:id` | Get single ad copy |
| POST | `/api/v1/ad-copies` | Create ad copy |
| PUT | `/api/v1/ad-copies/:id` | Update ad copy |
| DELETE | `/api/v1/ad-copies/:id` | Delete ad copy |

**Create Ad Copy:**
```json
{
  "name": "Hackathon Promo 1",
  "content": "🚀 Participating in a hackathon? Check out our dev tools!",
  "category": "hackathon",
  "priority": 10
}
```

## 📊 Workflow Process

```
1. Fetch all followed users
        ↓
2. Get latest N tweets for each user
        ↓
3. Call LLM to detect hackathon-related content
        ↓
4. Reply with ad copy on relevant tweets
        ↓
5. Log reply records
```

## 🛡️ Important Notes

1. **Twitter API Limits**: Be mindful of rate limits; set reasonable reply intervals
2. **Account Safety**: Avoid excessive auto-replies; recommended max 100 replies/day
3. **Content Compliance**: Ensure ad content complies with Twitter Terms of Service
4. **LLM Costs**: Each detection calls the LLM API; monitor usage costs

## 📄 License

MIT License

