package lock

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"math/rand"
	"time"
)

const (
	LockStatusOpen   = 0
	LockStatusClosed = 1
)

type MySQCounterLock struct {
	MySQLClient *gorm.DB
	ClientID    string
}

func NewCounterLockWithDSN(dsn string, debug bool) MySQLLockItf {
	l := MySQCounterLock{}
	l.MySQLClient = InitMySQLWithDSN(dsn, debug)
	id := uuid.New()
	l.ClientID = id.String()
	return &l
}

func NewCounterLockWithConn(db *sql.DB) MySQLLockItf {
	l := MySQCounterLock{}
	l.MySQLClient = InitMySQLWithConn(db)
	id := uuid.New()
	l.ClientID = id.String()
	return &l
}

// Init :Create tables
func (l *MySQCounterLock) Init(lockNameList []string) {
	if !l.MySQLClient.Migrator().HasTable(&LockCounter{}) {
		// Do not handle errors, because Init() can be executed multiple times
		_ = l.MySQLClient.Migrator().CreateTable(&LockCounter{})
	}

	for _, lockName := range lockNameList {
		var item LockCounter
		result := l.MySQLClient.Where("name = ?", lockName).Take(&item)
		if result.Error == gorm.ErrRecordNotFound {
			l.MySQLClient.Clauses(clause.OnConflict{DoNothing: true}).
				Create(&LockCounter{Name: lockName,
					Counter:    LockStatusOpen,
					Owner:      "",
					CreatedAt:  time.Now(),
					ModifiedAt: time.Now()})
		}
	}
}

func (l *MySQCounterLock) SetClientID(clientID string) {
	l.ClientID = clientID
}

// Lock :If the lock cannot be obtained, it will keep blocking
// wait: < 0 no wait
func (l *MySQCounterLock) Acquire(lockName string, wait time.Duration) error {
	mysqlClient := l.MySQLClient
	var record LockCounter

	mysqlClient.Where("name = ?", lockName).First(&record)
	if mysqlClient.Error != nil {
		return mysqlClient.Error
	}

	// reentrant lock
	if record.Counter == LockStatusClosed && record.Owner == l.ClientID {
		return nil
	}

	//  Lock is open.
	if record.Counter == LockStatusOpen {
		result := mysqlClient.Model(&LockCounter{}).Where("id = ? AND counter = ?",
			record.ID, LockStatusOpen).
			Updates(map[string]interface{}{
				"counter": LockStatusClosed,
				"owner":   l.ClientID},
			)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected >= 1 {
			return nil
		}
	}

	// retry
	start := time.Now()
	deadline := start.Add(wait)
	for time.Now().Before(deadline) {
		result := mysqlClient.Model(&LockCounter{}).Where("id = ? AND counter = ?",
			record.ID, LockStatusOpen).
			Updates(map[string]interface{}{
				"counter": LockStatusClosed,
				"owner":   l.ClientID},
			)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected >= 1 {
			return nil
		}

		time.Sleep(time.Duration(100+rand.Intn(100)) * time.Millisecond)
	}

	return fmt.Errorf("get lock timeout, wait:%v", wait)
}

func (l *MySQCounterLock) Release(lockName string) error {
	mysqlClient := l.MySQLClient
	var record LockCounter

	result := mysqlClient.Where("name = ?", lockName).First(&record)
	if result.Error != nil {
		return result.Error
	}

	if record.Counter == LockStatusOpen {
		return fmt.Errorf("lock [%v] is open", lockName)
	}

	if record.Owner != l.ClientID {
		return fmt.Errorf("lock [%v] is't owned by you", lockName)
	}

	result = mysqlClient.Model(&LockCounter{}).Where("id = ? AND owner = ?",
		record.ID, l.ClientID).
		Updates(map[string]interface{}{
			"counter": LockStatusOpen,
			"owner":   ""},
		)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
