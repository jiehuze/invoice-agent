package distributed

import (
	"sync"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

/*
	分布式锁 https://github.com/go-redsync/redsync
	*redsync.Mutex.Lock()
	*redsync.Mutex.UnLock()
*/
var (
	lockerOnce sync.Once
	locker     *redsync.Redsync
)

func GetDistributedLock(Name string, ops ...redsync.Option) *redsync.Mutex {
	return locker.NewMutex(Name, ops...)
}

func InitRedisLocker(host string, password string, db int) {
	lockerOnce.Do(func() {
		cli := redis.NewClient(&redis.Options{
			Addr:     host,
			Password: password,
			DB:       db,
		})
		pool := goredis.NewPool(cli)
		locker = redsync.New(pool)
	})
}
