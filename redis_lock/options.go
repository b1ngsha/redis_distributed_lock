package redis_lock

const (
	DefaultIdleTimeoutSeconds = 10
	DefaultMaxActive          = 100
	DefaultMaxIdle            = 20

	DefaultBlockWaitingSeconds = 5
	DefaultExpireSeconds       = 30
)

type ClientOptions struct {
	network  string
	address  string
	password string

	maxIdle            int
	idleTimeoutSeconds int
	maxActive          int
	wait               bool
}

type ClientOption func(c *ClientOptions)

func WithMaxIdle(maxIdle int) ClientOption {
	return func(c *ClientOptions) {
		c.maxIdle = maxIdle
	}
}

func WithIdleTimeoutSeconds(idleTimeoutSeconds int) ClientOption {
	return func(c *ClientOptions) {
		c.idleTimeoutSeconds = idleTimeoutSeconds
	}
}

func WithMaxActive(maxActive int) ClientOption {
	return func(c *ClientOptions) {
		c.maxActive = maxActive
	}
}

func WithWaitMode() ClientOption {
	return func(c *ClientOptions) {
		c.wait = true
	}
}

func repairClient(c *ClientOptions) {
	if c.maxIdle < 0 {
		c.maxIdle = DefaultMaxIdle
	}

	if c.idleTimeoutSeconds < 0 {
		c.idleTimeoutSeconds = DefaultIdleTimeoutSeconds
	}

	if c.maxActive < 0 {
		c.maxActive = DefaultMaxActive
	}
}

type LockOptions struct {
	isBlock             bool
	blockWaitingSeconds int64
	expireSeconds       int64
}

type LockOption func(*LockOptions)

func WithBlock() LockOption {
	return func(lo *LockOptions) {
		lo.isBlock = true
	}
}

func WithBlockWaitingSeconds(waitingSeconds int64) LockOption {
	return func(lo *LockOptions) {
		lo.blockWaitingSeconds = waitingSeconds
	}
}

func WithExpireSeconds(expireSeconds int64) LockOption {
	return func(lo *LockOptions) {
		lo.expireSeconds = expireSeconds
	}
}

func repairLock(lo *LockOptions) {
	if lo.isBlock && lo.blockWaitingSeconds <= 0 {
		lo.blockWaitingSeconds = DefaultBlockWaitingSeconds
	}

	if lo.expireSeconds <= 0 {
		lo.expireSeconds = DefaultExpireSeconds
	}
}
