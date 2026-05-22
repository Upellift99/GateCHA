//go:build mysql

package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {
	dialectors["mysql"] = dialectorEntry{
		open: func(dsn string) (gorm.Dialector, error) {
			return mysql.Open(dsn), nil
		},
		multiConn: true,
	}
}
