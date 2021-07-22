package lock

import (
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
	"time"
)

type MySQLLock struct {
	MySQLClient *gorm.DB
	TXMap       map[string]*gorm.DB
}

func NewMySQLLock(dsn string, debug bool) *MySQLLock {
	l := MySQLLock{}
	l.MySQLClient = InitMySQL(dsn, debug)
	l.TXMap = make(map[string]*gorm.DB)
	return &l
}

// Init :Create tables
// Create row record
func (l *MySQLLock) Init(lockNameList []string) {
	if !l.MySQLClient.Migrator().HasTable(&LockStore{}) {
		// Do not handle errors, because Init() can be executed multiple times
		_ = l.MySQLClient.Migrator().CreateTable(&LockStore{})
	}

	for _, lockName := range lockNameList {
		var item LockStore
		result := l.MySQLClient.Where("name = ?", lockName).Take(&item)
		if result.Error == gorm.ErrRecordNotFound {
			l.MySQLClient.Clauses(clause.OnConflict{DoNothing: true}).
				Create(&LockStore{Name: lockName, CreatedAt: time.Now()})
		}
	}
}

// Lock :If the lock cannot be obtained, it will keep blocking
func (l *MySQLLock) Acquire(lockName string, wait time.Duration) error {
	// start transaction
	tx, ok := l.TXMap[lockName]
	if !ok {
		l.TXMap[lockName] = l.MySQLClient.Begin()
		tx = l.TXMap[lockName]
	}
	beginTime := time.Now()
	var err error
	for time.Since(beginTime) < wait {
		result := tx.Exec("select * from _lock_store where name = ? for update", lockName)
		err = result.Error
		// Error 1205: Lock wait timeout exceeded; try restarting transaction
		if err == nil {
			break
		} else if strings.Index(err.Error(), "Error 1205") == -1 {
			// 如果成功获得锁，或者有其它错误，则退出
			break
		}
	}
	return err
}

func (l *MySQLLock) Release(lockName string) error {
	tx, ok := l.TXMap[lockName]
	if !ok {
		return fmt.Errorf("The lock must be locked before the lock can be released:%v", lockName)
	}
	tx.Commit()
	return nil
}
