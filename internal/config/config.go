package config

import (
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Twitter  TwitterConfig  `mapstructure:"twitter"`
	LLM      LLMConfig      `mapstructure:"llm"`
	Workflow WorkflowConfig `mapstructure:"workflow"`
	Log      LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type TwitterConfig struct {
	APIKey        string        `mapstructure:"api_key"`
	APISecret     string        `mapstructure:"api_secret"`
	AccessToken   string        `mapstructure:"access_token"`
	AccessSecret  string        `mapstructure:"access_secret"`
	BearerToken   string        `mapstructure:"bearer_token"`
	Timeout       time.Duration `mapstructure:"timeout"`
	RateLimitWait time.Duration `mapstructure:"rate_limit_wait"`
}

type LLMConfig struct {
	Provider   string        `mapstructure:"provider"`
	APIKey     string        `mapstructure:"api_key"`
	Model      string        `mapstructure:"model"`
	BaseURL    string        `mapstructure:"base_url"`
	Timeout    time.Duration `mapstructure:"timeout"`
	MaxRetries int           `mapstructure:"max_retries"`
}

type WorkflowConfig struct {
	DefaultTweetCount int           `mapstructure:"default_tweet_count"`
	ReplyInterval     time.Duration `mapstructure:"reply_interval"`
	MaxDailyReplies   int           `mapstructure:"max_daily_replies"`
	EnableScheduler   bool          `mapstructure:"enable_scheduler"`
	Schedule          string        `mapstructure:"schedule"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 支持环境变量
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// 处理环境变量替换
	cfg.resolveEnvVars()

	return &cfg, nil
}

func (c *Config) resolveEnvVars() {
	c.Database.Password = resolveEnv(c.Database.Password)
	c.Twitter.APIKey = resolveEnv(c.Twitter.APIKey)
	c.Twitter.APISecret = resolveEnv(c.Twitter.APISecret)
	c.Twitter.AccessToken = resolveEnv(c.Twitter.AccessToken)
	c.Twitter.AccessSecret = resolveEnv(c.Twitter.AccessSecret)
	c.Twitter.BearerToken = resolveEnv(c.Twitter.BearerToken)
	c.LLM.APIKey = resolveEnv(c.LLM.APIKey)
}

func resolveEnv(value string) string {
	if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
		envKey := value[2 : len(value)-1]
		return os.Getenv(envKey)
	}
	return value
}

