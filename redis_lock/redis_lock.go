package redis_lock

import (
	"errors"
	"redis_distributed_lock/utils"
)

var ErrLockAcquiredByOthers = errors.New("lock is acquired by others")

func IsAcquiredErr(err error) bool {
	return errors.Is(err, ErrLockAcquiredByOthers)
}

type RedisLock struct {
	LockOptions
	key    string
	token  string
	client *Client
}

func NewRedisLock(key string, client *Client, opts ...LockOption) *RedisLock {
	lock := RedisLock{
		key:    key,
		token:  utils.GetProcessAndGoroutineIDStr(),
		client: client,
	}

	for _, opt := range opts {
		opt(&lock.LockOptions)
	}

	repairLock(&lock.LockOptions)
	return &lock
}
