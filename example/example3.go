package main

import (
	"database/sql"
	mlock "github.com/vearne/mysql-lock"
	"log"
	"sync"
	"time"
)

func main() {
	dsn := "tc_user:20C462C9C614@tcp(127.0.0.1:3306)/xxx?charset=utf8&loc=Asia%2FShanghai&parseTime=true"
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	var locker mlock.MySQLLockItf
	locker = mlock.NewRowLockWithConn(sqlDB)
	//locker = mlock.NewCounterLockWithConn(sqlDB)

	locker.Init([]string{"lock1", "lock2"})
	// optional, only for CounterLock
	locker.SetClientID("client1")

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 2; i++ {
				beginTime := time.Now()
				err := locker.Acquire("lock1", 5*time.Second)
				if err != nil {
					log.Println("can't acquire lock", err)
					log.Println(time.Since(beginTime))
					return
				}

				log.Println("got lock1")
				log.Println(time.Since(beginTime))
				time.Sleep(3 * time.Second)
				locker.Release("lock1")
				log.Println("release lock1")
			}
		}()
	}

	wg.Wait()
}
