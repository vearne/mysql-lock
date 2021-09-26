# mysql-lock
基于MySQL的分布式锁

* [English README](https://github.com/vearne/mysql-lock/blob/master/README.md)

# 使用
```
go get github.com/vearne/mysql-lock
```

# 注意
* 基于mysql的分布式锁，是不严谨的。
  
  比如t1时刻，A持有锁，B等待加锁。t2时刻，A与MySQL之间的网络出现异常。MySQL主动释放了A所施加的锁(回滚了A没有提交的事务)，B加上了锁，这时候A会认为，它拥有锁；B也会认为自己持有锁。分布式锁其实失效了。
* mysql-lock会创建表 `_lock_store`。
  
  所以MySQL用户需要有`CREATE`权限。或者你可以使用 [doc/schema.sql](https://github.com/vearne/mysql-lock/blob/main/doc/schema.sql) 来自己创建表`_lock_store` .

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
	locker := mlock.NewMySQLLock(dsn, debug)
	// init() can be executed multiple times
	locker.Init([]string{"lock1", "lock2"})

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
