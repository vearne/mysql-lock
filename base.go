package lock

import (
	"time"
)

type MySQLLockItf interface {
	Init(lockNameList []string)
	SetClientID(clientID string)
	Acquire(lockName string, wait time.Duration) error
	Release(lockName string) error
}
