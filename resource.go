package lock

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

func InitMySQL(dsn string, debug bool) *gorm.DB {
	mysqldb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	if debug {
		mysqldb = mysqldb.Debug()
	}
	sqlDB, err := mysqldb.DB()
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(600 * time.Second)
	return mysqldb
}
