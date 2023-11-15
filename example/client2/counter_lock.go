package main

import (
	"context"
	mlock "github.com/vearne/mysql-lock"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	DSN = "root:9E68-2607F7855D7D@tcp(127.0.0.1:23406)/testdb?charset=utf8&loc=Asia%2FShanghai&parseTime=true"
)

func main() {
	debug := false
	//debug := true

	locker := mlock.NewLockClientWithDSN(DSN, debug)

	// init() can be executed multiple times
	locker.Init()

	clientID := "client2"
	lockName := "lock1"
	lock, err := locker.NewLock(lockName, mlock.WithClientID(clientID))
	if err != nil {
		log.Println("new lock", err)
		return
	}

	beginTime := time.Now()
	// max wait for 5 secs
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	err = lock.Acquire(ctx)
	if err != nil {
		log.Println("can't acquire lock", "error", err)
		log.Println(time.Since(beginTime))
		return
	}

	log.Printf("%s got lock [%s]\n", clientID, lockName)
	time.Sleep(10 * time.Second)
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
