-- 关注用户表
CREATE TABLE IF NOT EXISTS followed_users (
    id SERIAL PRIMARY KEY,
    twitter_user_id VARCHAR(64) NOT NULL UNIQUE,
    username VARCHAR(128) NOT NULL,
    display_name VARCHAR(256),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_followed_users_twitter_id ON followed_users(twitter_user_id);
CREATE INDEX IF NOT EXISTS idx_followed_users_active ON followed_users(is_active);

-- 广告文案表
CREATE TABLE IF NOT EXISTS ad_copies (
    id SERIAL PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    content TEXT NOT NULL,
    category VARCHAR(64) DEFAULT 'hackathon',
    priority INT DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    use_count INT DEFAULT 0,
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ad_copies_category ON ad_copies(category);
CREATE INDEX IF NOT EXISTS idx_ad_copies_active_priority ON ad_copies(is_active, priority DESC);

-- 回复日志表
CREATE TABLE IF NOT EXISTS reply_logs (
    id SERIAL PRIMARY KEY,
    tweet_id VARCHAR(64) NOT NULL,
    tweet_author_id VARCHAR(64) NOT NULL,
    tweet_content TEXT,
    reply_tweet_id VARCHAR(64),
    ad_copy_id INT REFERENCES ad_copies(id),
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    error_message TEXT,
    llm_response TEXT,
    is_hackathon BOOLEAN,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_reply_logs_tweet_id ON reply_logs(tweet_id);
CREATE INDEX IF NOT EXISTS idx_reply_logs_status ON reply_logs(status);
CREATE INDEX IF NOT EXISTS idx_reply_logs_created_at ON reply_logs(created_at);

-- 配置表
CREATE TABLE IF NOT EXISTS bot_configs (
    id SERIAL PRIMARY KEY,
    config_key VARCHAR(128) NOT NULL UNIQUE,
    config_value TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 插入默认配置
INSERT INTO bot_configs (config_key, config_value, description) VALUES
('default_tweet_count', '10', '默认获取推文数量'),
('reply_interval_seconds', '60', '回复间隔时间（秒）'),
('max_daily_replies', '100', '每日最大回复数')
ON CONFLICT (config_key) DO NOTHING;

-- 插入示例广告文案
INSERT INTO ad_copies (name, content, category, priority) VALUES
('黑客松推广1', '🚀 正在参加黑客松？来看看我们的开发者工具，助力你的项目脱颖而出！#Hackathon #Developer', 'hackathon', 10),
('黑客松推广2', '💡 黑客松参赛者必备！免费试用我们的API，让你的demo更出彩！', 'hackathon', 5)
ON CONFLICT DO NOTHING;

