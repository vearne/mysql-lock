package main

import (
	"context"
	mlock "github.com/vearne/mysql-lock"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	DSN = "root:9E68-2607F7855D7D@tcp(127.0.0.1:23406)/testdb?charset=utf8&loc=Asia%2FShanghai&parseTime=true"
)

// store a boolean variable
var isMaster atomic.Value

func main() {
	debug := false
	//debug := true

	locker := mlock.NewLockClientWithDSN(DSN, debug)

	// init() can be executed multiple times
	locker.Init()

	clientID := "client1"
	lockName := "lock1"
	lock, err := locker.NewLock(lockName,
		mlock.WithMaxLockTime(20*time.Second),
		mlock.WithClientID(clientID),
		mlock.WithCallBack(func(err error) {
			log.Println("refresh error", err)
			isMaster.Store(false)
			log.Println("##########################")
			log.Printf("%v is no longer master", clientID)
			log.Println("##########################")
		}),
	)
	if err != nil {
		log.Println("new lock", err)
		return
	}

	beginTime := time.Now()
	// wait until get lock
	err = lock.Acquire(context.Background())
	if err != nil {
		log.Println("can't acquire lock", "error", err)
		log.Println(time.Since(beginTime))
		return
	}

	log.Printf("%s got lock [%s]\n", clientID, lockName)
	isMaster.Store(true)

	beginTime = time.Now()
	for {
		if time.Now().Before(beginTime.Add(time.Minute * 10)) {
			time.Sleep(3 * time.Second)
			log.Println("isMaster:", isMaster.Load().(bool))
		}
	}

	err = lock.Release(context.Background())
	if err != nil {
		log.Printf("%s release lock [%s], error:%v\n", clientID, lockName, err)
	} else {
		log.Printf("%s release lock [%s]\n", clientID, lockName)
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	<-ch
}
