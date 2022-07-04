package main

import (
	mlock "github.com/vearne/mysql-lock"
	"log"
	"time"
)

func main() {
	debug := false
	//debug := true
	dsn := "tc_user:20C462C9C614@tcp(127.0.0.1:3306)/xxx?charset=utf8&loc=Asia%2FShanghai&parseTime=true"

	var locker *mlock.MySQCounterLock
	locker = mlock.NewCounterLockWithDSN(dsn, debug)

	// init() can be executed multiple times
	locker.Init([]string{"lock1", "lock2"})
	// optional, only for CounterLock
	//locker.SetClientID("client1")
	// optional, only for CounterLock
	// default 1 minutes
	locker.SetMaxLockTime(1 * time.Minute)
	// optional, only for CounterLock
	// default NewLinearBackOff(1*time.Second)
	locker.WithBackOff(mlock.NewLinearBackOff(2 * time.Second))

	beginTime := time.Now()
	// max wait for 5 secs
	err := locker.Acquire("lock1", 50*time.Second)
	if err != nil {
		log.Println("can't acquire lock", "error", err)
		log.Println(time.Since(beginTime))
		return
	}

	log.Println("got lock1")
	log.Println(time.Since(beginTime))
	time.Sleep(30 * time.Second)
	locker.Release("lock1")
	log.Println("release lock1")
}
