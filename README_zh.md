# mysql-lock

[![golang-ci](https://github.com/vearne/mysql-lock/actions/workflows/golang-ci.yml/badge.svg)](https://github.com/vearne/mysql-lock/actions/workflows/golang-ci.yml)

基于MySQL的分布式锁

* [English README](https://github.com/vearne/mysql-lock/blob/master/README.md)

## 用法
```
go get github.com/vearne/mysql-lock
```

## 初始化
```
mlock.NewCounterLockWithDSN()
mlock.NewCounterLockWithConn()
```

## 注意：基于mysql的分布式锁，是不严谨的:

比如t1时刻，A持有锁，B等待加锁。t2时刻，A与MySQL之间的网络出现异常。
MySQL自动释放了A所施加的锁，B加上了锁，
这时候A会认为，它拥有锁；B也会认为自己持有锁。分布式锁其实失效了。


## 注意：mysql-lock会创建表
* mysql-lock创建表`_lock_counter`。   

所以mysql-lock需要`CREATE`权限。 或者你可以使用 [doc/schema.sql](https://github.com/vearne/mysql-lock/blob/main/doc/schema.sql) 来自己创建表。

## 示例
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
	lock.Release(context.Background())
	log.Printf("%s release lock [%s]\n", clientID, lockName)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	<-ch
}
```
