package database

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func init() {
	dialectors["sqlite"] = dialectorEntry{
		open: func(dsn string) (gorm.Dialector, error) {
			dir := filepath.Dir(dsn)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create data directory: %w", err)
			}
			return sqlite.Open(dsn + "?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)"), nil
		},
		multiConn: false,
	}
}
