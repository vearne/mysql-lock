package main

import (
	mlock "github.com/vearne/mysql-lock"
	"log"
	"time"
)

func main() {
	debug := false
	dsn := "tc_user:D83B6CD42A6C@tcp(127.0.0.1:10104)/xxx?charset=utf8&loc=Asia%2FShanghai&parseTime=true"
	locker := mlock.NewMySQLLock(dsn, debug)
	locker.Init([]string{"lock1", "lock2"})

	for i := 0; i < 10; i++ {
		beginTime := time.Now()
		err := locker.Acquire("lock1", 5*time.Second)
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
