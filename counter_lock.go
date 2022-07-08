package lock

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	slog "github.com/vearne/simplelog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

const (
	LockStatusOpen   = 0
	LockStatusClosed = 1
)

type MySQLCounterLock struct {
	MySQLClient *gorm.DB
	ClientID    string
	MaxLockTime time.Duration
	// storage done channels
	store *DoneStore
	// backoff strategy
	backoff BackOff
}

func NewCounterLockWithDSN(dsn string, debug bool) *MySQLCounterLock {
	l := MySQLCounterLock{}
	l.MySQLClient = InitMySQLWithDSN(dsn, debug)
	l.ClientID = uuid.New().String()
	l.MaxLockTime = time.Minute
	l.store = NewDoneStore()
	l.backoff = NewLinearBackOff(time.Second)
	return &l
}

func NewCounterLockWithConn(db *sql.DB, debug bool) *MySQLCounterLock {
	l := MySQLCounterLock{}
	l.MySQLClient = InitMySQLWithConn(db, debug)
	l.ClientID = uuid.New().String()
	l.MaxLockTime = time.Minute
	l.store = NewDoneStore()
	l.backoff = NewLinearBackOff(time.Second)
	return &l
}

// Init :Create tables
func (l *MySQLCounterLock) Init(lockNameList []string) {
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
					ModifiedAt: time.Now(),
					ExpiredAt:  time.Now(),
				},
				)
		}
	}
}

func (l *MySQLCounterLock) SetClientID(clientID string) {
	l.ClientID = clientID
}

func (l *MySQLCounterLock) SetMaxLockTime(d time.Duration) {
	l.MaxLockTime = d
	if d < time.Minute {
		l.MaxLockTime = time.Minute
	}
}

func (l *MySQLCounterLock) WithBackOff(b BackOff) {
	l.backoff = b
}

func (l *MySQLCounterLock) Refresh(lockName string) error {
	ticker := time.NewTicker(l.MaxLockTime - time.Duration(10)*time.Second)
	defer ticker.Stop()
	var firstFlag bool = true

	slog.Debug("MySQCounterLock-Refresh worker starting...")
	for {
		select {
		case <-ticker.C:
			mysqlClient := l.MySQLClient
			result := mysqlClient.Model(&LockCounter{}).Where(
				"name = ? AND counter = ? AND owner = ?",
				lockName,
				LockStatusClosed,
				l.ClientID,
			).
				Updates(map[string]interface{}{
					"owner":      l.ClientID,
					"expired_at": time.Now().Add(l.MaxLockTime),
				},
				)
			if result.Error != nil {
				return result.Error
			}
			if firstFlag {
				firstFlag = false
				ticker.Reset(l.MaxLockTime)
			}
			slog.Debug("refresh lock")
		case <-l.store.GetDoneChan(lockName):
			return nil
		}
	}
}

func (l *MySQLCounterLock) StopRefresh(lockName string) {
	slog.Debug("StopRefresh")
	l.store.CLoseDoneChan(lockName)
}

// Lock :If the lock cannot be obtained, it will keep blocking
// wait: < 0 no wait
func (l *MySQLCounterLock) Acquire(lockName string, wait time.Duration) error {
	mysqlClient := l.MySQLClient
	var record LockCounter
	var deadline time.Time
	var start time.Time

	result := mysqlClient.Where("name = ?", lockName).First(&record)
	if result.Error != nil {
		return result.Error
	}

	// reentrant lock
	if record.Counter == LockStatusClosed && record.Owner == l.ClientID {
		return nil
	}

	//  Lock is open or Lock is expired.
	result = mysqlClient.Model(&LockCounter{}).Where(
		"id = ? AND ((counter = ?) or (expired_at <  NOW()))",
		record.ID, LockStatusOpen).
		Updates(map[string]interface{}{
			"counter":    LockStatusClosed,
			"owner":      l.ClientID,
			"expired_at": time.Now().Add(l.MaxLockTime),
		},
		)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected >= 1 {
		goto SUCESSGOT
	}

	// retry
	start = time.Now()
	deadline = start.Add(wait)
	for time.Now().Before(deadline) {
		slog.Debug("Acquire retry lock[%v]", lockName)

		result = mysqlClient.Model(&LockCounter{}).Where(
			"id = ? AND ((counter = ?) or (expired_at <  NOW()))",
			record.ID, LockStatusOpen).
			Updates(map[string]interface{}{
				"counter":    LockStatusClosed,
				"owner":      l.ClientID,
				"expired_at": time.Now().Add(l.MaxLockTime),
			},
			)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected >= 1 {
			goto SUCESSGOT
		}

		time.Sleep(l.backoff.NextBackOff())
	}

	return fmt.Errorf("get lock timeout, wait:%v", wait)

SUCESSGOT:
	//nolint:errcheck
	go l.Refresh(lockName)
	return nil
}

func (l *MySQLCounterLock) Release(lockName string) error {
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

	// stop refresh
	l.StopRefresh(lockName)
	return nil
}
