package redis_lock

import (
	"context"
	"errors"
	"fmt"
	"redis_distributed_lock/utils"
	"time"
)

var ErrLockAcquiredByOthers = errors.New("lock is acquired by others")

const (
	redisLockKeyPrefix = "REDIS_LOCK_"
	tickerDuration     = 50
    luaCheckAndDeleteLockScript = `
        local lockKey = KEYS[1]
        local lockToken = ARGV[1]
        local currentToken = redis.call('GET', lockKey)
        if (not currentToken or currentToken ~= lockToken) then
            return 0
        else
            return redis.call('DEL', lockKey)
        end
    `
)

func isRetryableErr(err error) bool {
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

func (lock *RedisLock) getLockKey() string {
	return redisLockKeyPrefix + lock.key
}

func (lock *RedisLock) tryLock(ctx context.Context) error {
	resp, err := lock.client.SetNX(ctx, lock.getLockKey(), lock.token, lock.expireSeconds)
	if err != nil {
		return err
	}
	if resp != 1 {
		return fmt.Errorf("resp: %d, err: %w", resp, ErrLockAcquiredByOthers)
	}
	return nil
}

func (lock *RedisLock) blockingLock(ctx context.Context) error {
	timeLimit := time.After(time.Duration(lock.blockWaitingSeconds) * time.Second)
	ticker := time.NewTicker(time.Duration(tickerDuration) * time.Millisecond)
    defer ticker.Stop()

    for range ticker.C {
        select {
        case <-ctx.Done():
            return fmt.Errorf("lock failed, ctx timeout, error: %w", ctx.Err())
        case <-timeLimit:
            return fmt.Errorf("block waiting timeout, error: %w", ErrLockAcquiredByOthers)
        default:
        }

        err := lock.tryLock(ctx)
        if err == nil {
            return nil
        }

        if !isRetryableErr(err) {
            return err
        }
    }

    return nil
}

func (lock *RedisLock) Lock(ctx context.Context) error {
	err := lock.tryLock(ctx)
	if err == nil {
		return nil
	}

	if !lock.isBlock {
		return err
	}

	if !isRetryableErr(err) {
		return err
	}

	return lock.blockingLock(ctx)
}

func (lock *RedisLock) Unlock(ctx context.Context) error {
    keysAndArgs := []interface{}{lock.getLockKey(), lock.token}
    res, err := lock.client.Eval(ctx, luaCheckAndDeleteLockScript, 1, keysAndArgs)
    if err != nil {
        return err
    }
    if ret, _ := res.(int64); ret != 1 {
        return errors.New("can not unlock without ownership of lock")
    }
    return nil
}
