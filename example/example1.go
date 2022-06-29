package main

import (
	mlock "github.com/vearne/mysql-lock"
	"log"
	"time"
)

func main() {
	debug := true
	dsn := "tc_user:20C462C9C614@tcp(127.0.0.1:3306)/xxx?charset=utf8&loc=Asia%2FShanghai&parseTime=true"
	var locker mlock.MySQLLockItf
	locker = mlock.NewRowLockWithDSN(dsn, debug)
	//locker = mlock.NewCounterLockWithDSN(dsn, debug)

	locker.Init([]string{"lock1", "lock2"})
	// optional, only for CounterLock
	locker.SetClientID("client1")

	for i := 0; i < 3; i++ {
		beginTime := time.Now()
		//err := locker.Acquire("lock1", 5*time.Second)
		err := locker.Acquire("lock1", -1)
		if err != nil {
			log.Println("can't acquire lock", err)
			log.Println(time.Since(beginTime))
			return
		}

		log.Println("got lock1")
		log.Println(time.Since(beginTime))
		time.Sleep(8 * time.Second)
		locker.Release("lock1")
		log.Println("release lock1")
	}

}
