package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zhoubofsy/x-bot/internal/application/service"
	"github.com/zhoubofsy/x-bot/internal/config"
	"github.com/zhoubofsy/x-bot/internal/infrastructure/llm"
	"github.com/zhoubofsy/x-bot/internal/infrastructure/persistence/postgres"
	"github.com/zhoubofsy/x-bot/internal/infrastructure/twitter"
	"github.com/zhoubofsy/x-bot/internal/interfaces/api"
	"github.com/zhoubofsy/x-bot/internal/interfaces/api/handler"
	"github.com/zhoubofsy/x-bot/internal/interfaces/scheduler"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	logger := initLogger(cfg.Log)
	defer logger.Sync()

	logger.Info("启动 X-Bot 服务",
		zap.String("config", *configPath),
		zap.String("mode", cfg.Server.Mode),
	)

	// 初始化数据库
	db, err := postgres.NewDB(&cfg.Database)
	if err != nil {
		logger.Fatal("连接数据库失败", zap.Error(err))
	}

	// 自动迁移（仅创建不存在的表，不修改已存在的约束）
	if err := postgres.MigrateWithoutConstraints(db); err != nil {
		logger.Fatal("数据库迁移失败", zap.Error(err))
	}

	// 初始化仓储
	userRepo := postgres.NewUserRepository(db)
	adCopyRepo := postgres.NewAdCopyRepository(db)
	replyLogRepo := postgres.NewReplyLogRepository(db)

	// 初始化外部客户端
	twitterClient := twitter.NewClient(&cfg.Twitter)
	llmClient := llm.NewClient(&cfg.LLM)

	// 初始化服务
	followerService := service.NewFollowerService(userRepo, twitterClient, logger)
	tweetService := service.NewTweetService(twitterClient, logger)
	hackathonDetector := service.NewHackathonDetector(llmClient, logger)
	adReplyService := service.NewAdReplyService(adCopyRepo, twitterClient, logger)
	workflowService := service.NewWorkflowService(
		followerService,
		tweetService,
		hackathonDetector,
		adReplyService,
		replyLogRepo,
		&cfg.Workflow,
		logger,
	)

	// 初始化 HTTP handlers
	workflowHandler := handler.NewWorkflowHandler(workflowService, followerService, replyLogRepo)
	adCopyHandler := handler.NewAdCopyHandler(adCopyRepo)

	// 初始化路由
	apiKey := os.Getenv("API_KEY")
	router := api.NewRouter(workflowHandler, adCopyHandler, cfg.Server.Mode, apiKey)

	// 初始化定时任务
	sched := scheduler.NewScheduler(workflowService, &cfg.Workflow, logger)
	if err := sched.Start(); err != nil {
		logger.Error("启动定时任务失败", zap.Error(err))
	}

	// 启动 HTTP 服务
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router.Engine(),
	}

	go func() {
		logger.Info("HTTP 服务启动", zap.Int("port", cfg.Server.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP 服务启动失败", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭服务...")

	// 停止定时任务
	sched.Stop()

	// 优雅关闭 HTTP 服务
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("HTTP 服务关闭失败", zap.Error(err))
	}

	logger.Info("服务已关闭")
}

func initLogger(cfg config.LogConfig) *zap.Logger {
	var level zapcore.Level
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	var zapCfg zap.Config
	if cfg.Format == "console" {
		zapCfg = zap.NewDevelopmentConfig()
	} else {
		zapCfg = zap.NewProductionConfig()
	}
	zapCfg.Level = zap.NewAtomicLevelAt(level)

	logger, _ := zapCfg.Build()
	return logger
}
