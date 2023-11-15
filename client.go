package lock

import (
	"database/sql"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

const (
	LockStatusOpen   = 0
	LockStatusClosed = 1
)

type LockClient struct {
	MySQLClient *gorm.DB
}

func NewLockClientWithDSN(dsn string, debug bool) *LockClient {
	l := LockClient{}
	l.MySQLClient = InitMySQLWithDSN(dsn, debug)
	return &l
}

func NewLockClientWithConn(db *sql.DB, debug bool) *LockClient {
	l := LockClient{}
	l.MySQLClient = InitMySQLWithConn(db, debug)
	return &l
}

// Init :Create tables
func (client *LockClient) Init() {
	if !client.MySQLClient.Migrator().HasTable(&LockCounter{}) {
		// Do not handle errors, because Init() can be executed multiple times
		_ = client.MySQLClient.Migrator().CreateTable(&LockCounter{})
	}
}

func (client *LockClient) NewLock(lockName string, opts ...Option) (*MySQLLock, error) {
	// create record
	item := LockCounter{Name: lockName,
		Counter:    LockStatusOpen,
		Owner:      "",
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
		ExpiredAt:  time.Now(),
	}
	err := client.MySQLClient.Where("name = ?", lockName).FirstOrCreate(&item).Error
	if err != nil {
		return nil, err
	}

	// init lock
	var lock MySQLLock
	lock.clientID = uuid.Must(uuid.NewUUID()).String()
	lock.name = lockName
	lock.mySQLClient = client.MySQLClient
	lock.backoff = NewLinearBackOff(time.Second)
	lock.maxLockTime = time.Second * 10
	lock.stopRefresh = make(chan struct{})
	lock.refreshBG = true

	// Loop through each option
	for _, opt := range opts {
		// Call the option giving the instantiated
		opt(&lock)
	}

	return &lock, nil
}
