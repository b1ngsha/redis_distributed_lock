package redis_lock

import "errors"

var LockAcquiredByOthersErr = errors.New("Lock is acquired by others")

func IsAcquiredErr(err error) bool {
	return errors.Is(err, LockAcquiredByOthersErr)
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
