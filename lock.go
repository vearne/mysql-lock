package lock

import (
	"context"
	"errors"
	"fmt"
	slog "github.com/vearne/simplelog"
	"gorm.io/gorm"
	"time"
)

var (
	AcquireTimeoutErr error = errors.New("get lock timeout")
)

type CallBackFunc func(err error)

type MySQLLock struct {
	name        string
	clientID    string
	mySQLClient *gorm.DB
	maxLockTime time.Duration
	backoff     BackOff
	stopRefresh chan struct{}
	// Perform refresh in background
	refreshBG bool
	// If an error occurs when performing refresh, execute this callback
	callback func(err error)
}

func (l *MySQLLock) Acquire(ctx context.Context) error {
	mysqlClient := l.mySQLClient
	var record LockCounter

	result := mysqlClient.Where("name = ?", l.name).First(&record)
	if result.Error != nil {
		return result.Error
	}

	// reentrant lock
	if record.Counter == LockStatusClosed && record.Owner == l.clientID &&
		record.ExpiredAt.After(time.Now().Add(time.Second+l.maxLockTime/2)) {
		goto SUCESSGOT
	}

	//  Lock is open or Lock is expired.
	result = mysqlClient.Model(&LockCounter{}).Where(
		"id = ? AND ((counter = ?) or (expired_at <  NOW()))",
		record.ID, LockStatusOpen).
		Updates(map[string]interface{}{
			"counter":    LockStatusClosed,
			"owner":      l.clientID,
			"expired_at": time.Now().Add(l.maxLockTime),
		},
		)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected >= 1 {
		goto SUCESSGOT
	}

	// retry
	for {
		timer := time.NewTimer(l.backoff.NextBackOff())
		select {
		case <-ctx.Done():
			return AcquireTimeoutErr
		case <-timer.C:
			slog.Debug("Acquire retry lock[%v]", l.name)
			result = mysqlClient.Model(&LockCounter{}).Where(
				"id = ? AND ((counter = ?) or (expired_at <  NOW()))",
				record.ID, LockStatusOpen).
				Updates(map[string]interface{}{
					"counter":    LockStatusClosed,
					"owner":      l.clientID,
					"expired_at": time.Now().Add(l.maxLockTime),
				})
			if result.Error != nil {
				return result.Error
			}

			if result.RowsAffected >= 1 {
				goto SUCESSGOT
			}
		}
	}
SUCESSGOT:
	l.backoff.Reset()
	//nolint:errcheck
	if l.refreshBG {
		go func() {
			err := l.Refresh()
			if err != nil {
				slog.Error("refresh error, %v", err)
				if l.callback != nil {
					l.callback(err)
				}
			}
		}()
	}
	return nil
}

func (l *MySQLLock) Refresh() error {
	ticker := time.NewTicker(l.maxLockTime / 2)
	defer ticker.Stop()

	slog.Debug("MySQLLock-Refresh worker starting...")
	for {
		select {
		case <-ticker.C:
			mysqlClient := l.mySQLClient
			result := mysqlClient.Model(&LockCounter{}).Where(
				"name = ? AND counter = ? AND owner = ?",
				l.name,
				LockStatusClosed,
				l.clientID,
			).Updates(map[string]interface{}{
				"expired_at": time.Now().Add(l.maxLockTime),
			})
			if result.Error != nil {
				return result.Error
			}
			slog.Debug("refresh lock")
		case <-l.stopRefresh:
			slog.Debug("MySQLLock-Refresh stop")
			return nil
		}
	}
}

func (l *MySQLLock) Release(ctx context.Context) error {
	mysqlClient := l.mySQLClient
	var record LockCounter

	result := mysqlClient.Where("name = ?", l.name).First(&record)
	if result.Error != nil {
		return result.Error
	}

	if record.Counter == LockStatusOpen {
		return fmt.Errorf("lock [%v] is open", l.name)
	}

	if record.Owner != l.clientID {
		return fmt.Errorf("lock [%v] is't owned by you", l.name)
	}

	result = mysqlClient.Model(&LockCounter{}).Where(
		"id = ? AND owner = ?", record.ID, l.clientID).
		Updates(map[string]interface{}{
			"counter": LockStatusOpen,
			"owner":   "",
		},
		)
	if result.Error != nil {
		return result.Error
	}

	// stop refresh
	close(l.stopRefresh)
	return nil
}
