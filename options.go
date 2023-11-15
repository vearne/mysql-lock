package lock

import "time"

type Option func(lock *MySQLLock)

func WithBackOff(b BackOff) Option {
	return func(lock *MySQLLock) {
		lock.backoff = b
	}
}

func WithMaxLockTime(d time.Duration) Option {
	return func(lock *MySQLLock) {
		if d < 5*time.Second {
			d = 5 * time.Second
		}
		lock.maxLockTime = d
	}
}

func WithClientID(clientID string) Option {
	return func(lock *MySQLLock) {
		lock.clientID = clientID
	}
}

func WithRefreshBG(refreshBG bool) Option {
	return func(lock *MySQLLock) {
		lock.refreshBG = refreshBG
	}
}

// If an error occurs when performing refresh, execute this callback
func WithCallBack(f CallBackFunc) Option {
	return func(lock *MySQLLock) {
		lock.callback = f
	}
}
