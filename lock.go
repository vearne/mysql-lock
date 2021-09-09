package lock

import (
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

	if wait < 1*time.Second {
		wait = 1 * time.Second
	}
	var err error
	// The length of time in seconds an InnoDB transaction waits for a row lock before giving up.
	result := tx.Exec("SET @@session.innodb_lock_wait_timeout = ?", wait/time.Second)
	err = result.Error
	if err != nil {
		return err
	}

	result = tx.Exec("select * from _lock_store where name = ? for update", lockName)
	err = result.Error
	// Error 1205: Lock wait timeout exceeded; try restarting transaction
	return err
}

func (l *MySQLLock) Release(lockName string) error {
	tx, ok := l.TXMap[lockName]
	if !ok {
		return fmt.Errorf("The lock must be locked before the lock can be released:%v", lockName)
	}
	tx.Commit()
	// 注意: 每个事务都有唯一的事务ID
	// 当事务被commit或者rollback之后，就不能再使用了
	delete(l.TXMap, lockName)
	return nil
}
