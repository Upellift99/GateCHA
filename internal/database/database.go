package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open opens a database connection for the given driver and DSN and runs migrations.
//
// driver: "sqlite" or "mysql"
// dsn:    file path for SQLite; connection string for MySQL
//
// MySQL DSN example: "user:pass@tcp(host:3306)/dbname?parseTime=true&charset=utf8mb4&loc=UTC"
func Open(driver, dsn string) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch driver {
	case "sqlite":
		dir := filepath.Dir(dsn)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create data directory: %w", err)
		}
		// Append WAL mode and foreign key enforcement pragmas.
		dialector = sqlite.Open(dsn + "?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)")
	case "mysql":
		dialector = mysql.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported DB driver %q: must be \"sqlite\" or \"mysql\"", driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if driver == "sqlite" {
		sqlDB.SetMaxOpenConns(1)
	} else {
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
