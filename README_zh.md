# mysql-lock

[![golang-ci](https://github.com/vearne/mysql-lock/actions/workflows/golang-ci.yml/badge.svg)](https://github.com/vearne/mysql-lock/actions/workflows/golang-ci.yml)

基于MySQL的分布式锁

* [English README](https://github.com/vearne/mysql-lock/blob/master/README.md)

# 用法
```
go get github.com/vearne/mysql-lock
```


mysql-lock 使用2种方法创建MySQL分布式锁

* 方法1：行锁

 初始化
```
mlock.NewRowLockWithDSN()
mlock.NewRowLockWithConn()
```
* 方法2：设置标识

初始化
```
mlock.NewCounterLockWithDSN()
mlock.NewCounterLockWithConn()
```


# 注意
**基于mysql的分布式锁，是不严谨的。**


## 对于方法1:
比如t1时刻，A持有锁，B等待加锁。t2时刻，A与MySQL之间的网络出现异常。MySQL主动释放了A所施加的锁(回滚了A没有提交的事务)，B加上了锁，这时候A会认为，它拥有锁；B也会认为自己持有锁。分布式锁其实失效了。


## 对于方法2
比如t1时刻，A持有锁，B等待加锁。t2时刻，A突然崩溃，导致没有正常释放锁。那么B将永远无法获得锁


mysql-lock会创建表 方法1创建表`_lock_store`。方法2创建表`_lock_counter`。 所以mysql-lock需要`CREATE`权限。或者你可以使用 [doc/schema.sql](https://github.com/vearne/mysql-lock/blob/main/doc/schema.sql) 来自己创建表。

# 示例
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
