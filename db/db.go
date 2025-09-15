package db

import (
	"time"

	"github.com/coconutLatte/ezchat-server/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB *gorm.DB
)

type Options struct {
	DSN            string
	MaxOpenConns   int
	MaxIdleConns   int
	ConnMaxIdleSec int
	ConnMaxLifeSec int
}

func Init(opts Options) error {
	gdb, err := gorm.Open(mysql.Open(opts.DSN), &gorm.Config{})
	if err != nil {
		return err
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		return err
	}
	if opts.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(opts.MaxOpenConns)
	}
	if opts.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(opts.MaxIdleConns)
	}
	if opts.ConnMaxIdleSec > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(opts.ConnMaxIdleSec) * time.Second)
	}
	if opts.ConnMaxLifeSec > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(opts.ConnMaxLifeSec) * time.Second)
	}

	// ping
	if err := sqlDB.Ping(); err != nil {
		logger.S().Error("failed to ping database", "error", err)
		return err
	}

	DB = gdb
	return nil
}
