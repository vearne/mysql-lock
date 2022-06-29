# mysql-lock

[![golang-ci](https://github.com/vearne/mysql-lock/actions/workflows/golang-ci.yml/badge.svg)](https://github.com/vearne/mysql-lock/actions/workflows/golang-ci.yml)

a simple distributed lock based on mysql

* [中文 README](https://github.com/vearne/mysql-lock/blob/master/README_zh.md)

# Usage
```
go get github.com/vearne/mysql-lock
```


mysql-lock uses 2 methods to build MySQL's distributed lock

* Method 1: row lock
  initialization
```
mlock.NewRowLockWithDSN()
mlock.NewRowLockWithConn()
```
* Method 2: set flag
  initialization
```
mlock.NewCounterLockWithDSN()
mlock.NewCounterLockWithConn()
```


# Notice
**Distributed locks based on mysql are not rigorous.**


* Method 1:
  For example, at t1, A holds the lock, and B waits for the lock. At t2, the network between A and MySQL is abnormal. MySQL actively releases the lock imposed by A (rolling back the transaction that A has not committed), and B adds the lock. At this time, A will think that it owns the lock; B will also think that it holds the lock. The distributed lock actually fails.


* Method 2:
  For example, at time t1, A holds the lock and B waits for the lock. At time t2, A suddenly crashes, causing the lock to not be released normally. then B will never be able to acquire the lock.


mysql-lock will create the table. Method 1 creates the table `_lock_store`. Method 2 creates the table `_lock_counter`. So mysql-lock needs `CREATE` permission. Or you can use [doc/schema.sql](https://github.com/vearne/mysql-lock/blob/main/doc/schema.sql) to create the table yourself.

# Example
```
package main

import (
	mlock "github.com/vearne/mysql-lock"
	"log"
	"time"
)

func main() {
	debug := false
	dsn := "tc_user:20C462C9C614@tcp(127.0.0.1:3306)/xxx?charset=utf8&loc=Asia%2FShanghai&parseTime=true"

	var locker mlock.MySQLLockItf
	locker = mlock.NewRowLockWithDSN(dsn, debug)
	//locker = mlock.NewCounterLockWithDSN(dsn, debug)

	// init() can be executed multiple times
	locker.Init([]string{"lock1", "lock2"})
	
	// optional, only for CounterLock
	locker.SetClientID("client1")

	beginTime := time.Now()
	// max wait for 5 secs
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
```
