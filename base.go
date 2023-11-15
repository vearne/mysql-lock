package lock

import (
	"context"
	slog "github.com/vearne/simplelog"
	"os"
	"time"
)

type MySQLLockItf interface {
	Init(lockNameList []string)
	Acquire(lockName string, wait time.Duration) error
	Release(lockName string) error
}

type LockItf interface {
	Acquire(ctx context.Context) error
	Release(ctx context.Context) error
}

func init() {
	// export ENABLE_LOG=true
	flag := os.Getenv("ENABLE_LOG")
	if len(flag) > 0 {
		slog.SetOutput(os.Stdout)
	}

	// export LOG_LEVEL=debug
	level := os.Getenv("LOG_LEVEL")
	if len(level) > 0 {
		slog.SetLevel(slog.LogMap[level])
	} else {
		slog.SetLevel(slog.WarnLevel)
	}
}
