package lock

import (
	"database/sql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

// initialize *gorm.DB with dsn
func InitMySQLWithDSN(dsn string, debug bool) *gorm.DB {
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

// initialize *gorm.DB with an existing database connection
func InitMySQLWithConn(sqlDB *sql.DB, debug bool) *gorm.DB {
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	if debug {
		gormDB = gormDB.Debug()
	}
	return gormDB
}
