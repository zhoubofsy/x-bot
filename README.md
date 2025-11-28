# X-Bot: 黑客松推文广告自动回复系统

一个基于 Go 的自动化 Twitter 机器人，用于监控关注用户的推文，识别与"黑客松"相关的内容，并自动回复预制的广告文案。

## ✨ 功能特性

- **自动监控**: 获取关注用户的最新推文
- **智能识别**: 使用 LLM (GPT) 判断推文是否与黑客松相关
- **自动回复**: 在相关推文下自动回复预设的广告文案
- **广告管理**: 支持多广告文案管理，按优先级轮换
- **定时任务**: 支持 Cron 定时执行工作流
- **统计分析**: 回复日志记录与统计
- **API 接口**: RESTful API 支持手动触发和管理

## 🏗️ 技术架构

```
├── cmd/server/          # 程序入口
├── config/              # 配置文件
├── migrations/          # 数据库迁移脚本
├── internal/
│   ├── config/          # 配置加载
│   ├── domain/          # 领域层 (实体 + 仓储接口)
│   ├── infrastructure/  # 基础设施层 (DB/Twitter/LLM)
│   ├── application/     # 应用服务层 (业务逻辑)
│   └── interfaces/      # 接口层 (HTTP API + 定时任务)
└── pkg/                 # 公共工具包
```

**技术栈:**
- Go 1.21+
- PostgreSQL
- Gin (HTTP 框架)
- GORM (ORM)
- Twitter API v2 (OAuth 1.0a)
- OpenAI API (LLM)

## 🚀 快速开始

### 1. 环境要求

- Go 1.21+
- PostgreSQL 12+
- Twitter Developer Account (API v2 权限)
- OpenAI API Key

### 2. 克隆项目

```bash
git clone https://github.com/zhoubofsy/x-bot.git
cd x-bot
```

### 3. 配置环境变量

```bash
# 数据库
export DB_PASSWORD=your_db_password

# Twitter API (OAuth 1.0a)
export TWITTER_API_KEY=your_api_key
export TWITTER_API_SECRET=your_api_secret
export TWITTER_ACCESS_TOKEN=your_access_token
export TWITTER_ACCESS_SECRET=your_access_secret
export TWITTER_BEARER_TOKEN=your_bearer_token

# OpenAI
export OPENAI_API_KEY=your_openai_api_key

# API 认证 (可选)
export API_KEY=your_api_key_for_http_auth
```

### 4. 初始化数据库

```bash
# 创建数据库
createdb x_bot

# 执行迁移脚本
psql -U postgres -d x_bot -f migrations/001_init.sql
```

### 5. 编译运行

```bash
# 安装依赖
go mod tidy

# 编译
go build -o bin/x-bot ./cmd/server/

# 运行
./bin/x-bot -config config/config.yaml
```

## 📝 配置说明

编辑 `config/config.yaml`:

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
  default_tweet_count: 10      # 每用户获取推文数
  reply_interval: 60s          # 回复间隔
  max_daily_replies: 100       # 每日最大回复数
  enable_scheduler: true       # 启用定时任务
  schedule: "0 */2 * * *"      # Cron 表达式 (每2小时)
```

## 🔌 API 接口

所有 API 需要在 Header 中携带认证:
```
Authorization: Bearer <API_KEY>
```

### 工作流

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/workflow/execute` | 执行工作流 |
| POST | `/api/v1/workflow/sync-following` | 同步关注列表 |

**执行工作流参数:**
```json
{
  "tweet_count": 10,
  "dry_run": false
}
```

### 统计与日志

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/stats` | 获取统计信息 |
| GET | `/api/v1/reply-logs?limit=20` | 获取回复日志 |

### 广告文案管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/ad-copies` | 获取所有广告文案 |
| GET | `/api/v1/ad-copies/:id` | 获取单个广告文案 |
| POST | `/api/v1/ad-copies` | 创建广告文案 |
| PUT | `/api/v1/ad-copies/:id` | 更新广告文案 |
| DELETE | `/api/v1/ad-copies/:id` | 删除广告文案 |

**创建广告文案:**
```json
{
  "name": "黑客松推广1",
  "content": "🚀 正在参加黑客松？来看看我们的开发者工具！",
  "category": "hackathon",
  "priority": 10
}
```

## 📊 工作流程

```
1. 获取所有关注用户
        ↓
2. 获取每个用户的最新 N 条推文
        ↓
3. 调用 LLM 判断推文是否与黑客松相关
        ↓
4. 对相关推文回复广告文案
        ↓
5. 记录回复日志
```

## 🛡️ 注意事项

1. **Twitter API 限制**: 注意 API 速率限制，建议设置合理的回复间隔
2. **防止封号**: 避免过于频繁的自动回复，建议每日回复数不超过 100
3. **广告内容**: 确保广告内容符合 Twitter 使用条款
4. **LLM 成本**: 每次检测会调用 LLM API，注意控制成本

## 📄 License

MIT License

