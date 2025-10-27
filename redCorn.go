package redCorn

import (
	"context"
	"errors"
	"fmt"
	"time"

	goredislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"github.com/robfig/cron/v3"
)

// Cfg 配置结构体
type Cfg struct {
	RedisCfg goredislib.UniversalOptions
	LockCfg  LockCfg
	Logger   Logger // 自定义日志器，可选
}

type LockCfg struct {
	Expiry time.Duration
	Prefix string
}

// DistributedTaskManager 分布式任务管理器
type DistributedTaskManager struct {
	redisClient goredislib.UniversalClient
	redsync     *redsync.Redsync
	cron        *cron.Cron
	ctx         context.Context
	cancel      context.CancelFunc
	cfg         Cfg
	log         Logger
}

// NewDistributedTaskManager 创建分布式任务管理器
func NewDistributedTaskManager(cfg Cfg) (*DistributedTaskManager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// 设置日志器
	logger := cfg.Logger
	if logger == nil {
		logger = newDefaultLogger()
	}

	// 创建Redis客户端
	client := goredislib.NewUniversalClient(&cfg.RedisCfg)

	// 测试Redis连接
	if err := client.Ping(ctx).Err(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	// 创建Redsync连接池
	pool := goredis.NewPool(client)
	rs := redsync.New(pool)
	// 创建Cron实例
	c := cron.New(cron.WithSeconds()) // 支持秒级定时

	return &DistributedTaskManager{
		redisClient: client,
		redsync:     rs,
		cron:        c,
		ctx:         ctx,
		cancel:      cancel,
		cfg:         cfg,
		log:         logger,
	}, nil
}

// addDistributedTask 添加分布式定时任务
func (dtm *DistributedTaskManager) addDistributedTask(name, spec string, task func()) error {
	// 包装任务，添加分布式锁逻辑
	wrappedTask := func() {
		dtm.executeDistributedTask(name, task)
	}

	// 添加定时任务
	_, err := dtm.cron.AddFunc(spec, wrappedTask)
	if err != nil {
		return fmt.Errorf("failed to add cron task %s: %v", name, err)
	}

	dtm.log.Info("Added distributed task: ", name, ", schedule: ", spec)
	return nil
}

// executeDistributedTask 执行分布式任务（带锁）
func (dtm *DistributedTaskManager) executeDistributedTask(taskName string, task func()) {
	lockName := dtm.cfg.LockCfg.Prefix + taskName
	mutex := dtm.redsync.NewMutex(lockName, redsync.WithExpiry(dtm.cfg.LockCfg.Expiry))

	// 尝试获取分布式锁
	if err := mutex.TryLock(); err != nil {
		if errors.Is(err, redsync.ErrFailed) {
			dtm.log.Info("Task ", taskName, ": is running, skipping execution")
		} else {
			dtm.log.Error("Task ", taskName, ": Failed to acquire lock, skipping execution, err:", err)
		}
		return
	}

	// 确保释放锁
	defer func() {
		if ok, err := mutex.Unlock(); !ok || err != nil {
			if errors.Is(err, redsync.ErrLockAlreadyExpired) {
				dtm.log.Warn("WARN!!! Task ", taskName, ": LockCfg already expired, skipping release")
			} else {
				dtm.log.Error("Task ", taskName, ": Failed to release lock: ", err)
			}
		} else {
			dtm.log.Info("Task ", taskName, ": LockCfg released successfully")
		}
	}()

	dtm.log.Info("Task ", taskName, ": LockCfg acquired, starting execution")

	// 执行任务
	startTime := time.Now()
	task()
	duration := time.Since(startTime)

	dtm.log.Info("Task ", taskName, ": Completed in ", duration)
}

// Start 启动任务管理器
func (dtm *DistributedTaskManager) Start() {
	dtm.cron.Start()
	dtm.log.Info("Distributed task manager started")
}

// Stop 停止任务管理器
func (dtm *DistributedTaskManager) Stop() {
	dtm.log.Info("Stopping distributed task manager...")

	// 停止定时器
	ctx := dtm.cron.Stop()
	<-ctx.Done()

	// 取消上下文
	dtm.cancel()

	// 等待所有任务完成
	//dtm.wg.Wait()

	// 关闭Redis连接
	if err := dtm.redisClient.Close(); err != nil {
		dtm.log.Error("Error closing RedisCfg connection: ", err)
	}

	dtm.log.Info("Distributed task manager stopped")
}

// GetRedisClient 获取Redis客户端（供外部使用）
func (dtm *DistributedTaskManager) GetRedisClient() goredislib.UniversalClient {
	return dtm.redisClient
}

// GetContext 获取上下文（供外部使用）
func (dtm *DistributedTaskManager) GetContext() context.Context {
	return dtm.ctx
}

// AddScheduler 批量添加任务调度器中的所有任务
func (dtm *DistributedTaskManager) AddScheduler(scheduler *TaskScheduler) error {
	for name, schedule := range scheduler.GetAll() {
		if err := dtm.addDistributedTask(name, schedule.Cron, schedule.Task); err != nil {
			return fmt.Errorf("failed to add task %s: %v", name, err)
		}
	}
	return nil
}

// AddTask 仍然支持单个任务添加（保持灵活性）
func (dtm *DistributedTaskManager) AddTask(name, cron string, task func()) error {
	return dtm.addDistributedTask(name, cron, task)
}
