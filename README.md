# mysql-lock

[![golang-ci](https://github.com/vearne/mysql-lock/actions/workflows/golang-ci.yml/badge.svg)](https://github.com/vearne/mysql-lock/actions/workflows/golang-ci.yml)

a simple distributed lock based on mysql

* [中文 README](https://github.com/vearne/mysql-lock/blob/master/README_zh.md)

## Usage
```
go get github.com/vearne/mysql-lock
```

## initialization
```
mlock.NewCounterLockWithDSN()
mlock.NewCounterLockWithConn()
```


## Notice :Distributed locks based on mysql are not rigorous 
For example, at t1, A holds the lock, and B waits for the lock. 
At t2, the network between A and MySQL is abnormal. 
MySQL actively releases the lock imposed by A.And B adds the lock.
At this time, A will think that it owns the lock; B will also think that it holds the lock. 
The distributed lock actually fails.


## Notice: mysql-lock will create the table.
* `mysql-lock` creates the table `_lock_counter`.

So `mysql-lock` needs `CREATE` permission. Or you can use [doc/schema.sql](https://github.com/vearne/mysql-lock/blob/main/doc/schema.sql) to create the table yourself.

## Example
```
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

	clientID := "client1"
	lockName := "lock1"
	lock, err := locker.NewLock(lockName,
		mlock.WithMaxLockTime(20*time.Second),
		mlock.WithClientID(clientID))
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
	time.Sleep(10 * time.Second)
	lock.Release(context.Background())
	log.Printf("%s release lock [%s]\n", clientID, lockName)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	<-ch
}
```
