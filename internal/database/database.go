package database

import (
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// dialectorEntry holds a factory function for a GORM dialector and a flag
// indicating whether the driver supports multiple concurrent connections.
type dialectorEntry struct {
	open      func(dsn string) (gorm.Dialector, error)
	multiConn bool
}

// dialectors is populated by init() functions in the per-dialect files
// (dialect_sqlite.go, dialect_mysql.go, …).
var dialectors = map[string]dialectorEntry{}

// Open opens a database connection for the given driver and DSN.
//
// Which drivers are available depends on the build tags used at compile time:
//   - "sqlite" is always available
//   - "mysql"  is available when built with -tags mysql
//
// MySQL DSN example: "user:pass@tcp(host:3306)/dbname?parseTime=true&charset=utf8mb4&loc=UTC"
func Open(driver, dsn string) (*gorm.DB, error) {
	entry, ok := dialectors[driver]
	if !ok {
		supported := make([]string, 0, len(dialectors))
		for k := range dialectors {
			supported = append(supported, k)
		}
		sort.Strings(supported)
		return nil, fmt.Errorf("unsupported DB driver %q (built-in drivers: %v)", driver, supported)
	}

	dialector, err := entry.open(dsn)
	if err != nil {
		return nil, err
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

	if entry.multiConn {
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
	} else {
		sqlDB.SetMaxOpenConns(1)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
