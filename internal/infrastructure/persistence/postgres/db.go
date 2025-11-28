package postgres

import (
	"fmt"

	"github.com/zhoubofsy/x-bot/internal/config"
	"github.com/zhoubofsy/x-bot/internal/domain/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	logLevel := logger.Silent
	if cfg.SSLMode == "disable" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	// 使用 DisableForeignKeyConstraintWhenMigrating 避免约束问题
	return db.Set("gorm:table_options", "").
		AutoMigrate(
			&entity.FollowedUser{},
			&entity.AdCopy{},
			&entity.ReplyLog{},
			&entity.BotConfig{},
		)
}

// MigrateWithoutConstraints 跳过约束检查的迁移
// 适用于表已存在但约束不同的情况
func MigrateWithoutConstraints(db *gorm.DB) error {
	migrator := db.Migrator()

	tables := []interface{}{
		&entity.FollowedUser{},
		&entity.AdCopy{},
		&entity.ReplyLog{},
		&entity.BotConfig{},
	}

	for _, table := range tables {
		if !migrator.HasTable(table) {
			if err := migrator.CreateTable(table); err != nil {
				return err
			}
		}
	}

	return nil
}
