# 🌽 RedCorn - 分布式定时任务管理器

RedCorn 是一个基于 Redis 的分布式定时任务管理库，支持在多个节点间协调执行定时任务，确保任务在同一时间只有一个实例运行。

## ✨ 特性

- 🎯 **分布式协调** - 基于 Redis Redlock 算法，确保任务在集群中只有一个实例执行
- ⏰ **Cron 表达式** - 支持标准 Cron 表达式，灵活配置任务执行时间
- 🔄 **自动重试** - 内置重试机制，处理临时性故障
- 📝 **多种使用方式** - 支持直接添加任务、TaskScheduler 管理、批量添加
- 🔧 **可扩展** - 支持自定义日志器、配置灵活
- 🚀 **高性能** - 基于 go-redis 和 redsync，性能优异

## 📦 安装

```bash
go get github.com/kzdgt/redCorn
```

## 🚀 快速开始

### 方式一：直接添加任务（最简单）

```go
package main

import (
    "log"
    "time"
    "github.com/kzdgt/redCorn"
)

func main() {
    // 创建配置
    cfg := redCorn.Cfg{
        RedisCfg: redCorn.RedisCfg{
            Addr:     "localhost:6379",
            Password: "",
            DB:       0,
        },
        LockCfg: redCorn.LockCfg{
            Prefix: "myapp:lock:",
            Expiry: 30 * time.Second,
        },
    }
    
    // 创建任务管理器
    dtm, err := redCorn.NewDistributedTaskManager(cfg)
    if err != nil {
        log.Fatal(err)
    }

    // 直接添加定时任务
    dtm.AddTask("my-job", "*/10 * * * * *", func() {
        log.Println("每10秒执行一次任务...")
    })
    
    dtm.Start()
    defer dtm.Stop()
    
    select {} // 阻塞运行
}
```

### 方式二：使用 TaskScheduler（推荐）

```go
// 创建配置
cfg := redCorn.Cfg{
    RedisCfg: redCorn.RedisCfg{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
    },
    LockCfg: redCorn.LockCfg{
        Prefix: "myapp:lock:",
        Expiry: 30 * time.Second,
    },
}

// 创建任务管理器
dtm, err := redCorn.NewDistributedTaskManager(cfg)
if err != nil {
    log.Fatal(err)
}

// 创建任务调度器
scheduler := redCorn.NewTaskScheduler()

// 注册多个任务
scheduler.Register("health-check", "*/30 * * * * *", healthCheckTask)
scheduler.Register("data-sync", "0 */5 * * * *", dataSyncTask)
scheduler.Register("report", "0 0 2 * * *", reportTask)

// 批量添加到任务管理器
dtm.AddScheduler(scheduler)
```

### 方式三：单独添加调度任务

```go
if task, exists := scheduler.Get("data-sync"); exists {
    dtm.AddTask("sync-job", task.Cron, task.Task)
}
```

## ⚙️ 配置

### 环境变量配置

```bash
export REDIS_ADDR="localhost:6379"        # Redis 地址
export REDIS_PASSWORD="yourpassword"        # Redis 密码
export REDIS_DB="0"                         # Redis 数据库
export LOCK_EXPIRY="30s"                    # 锁过期时间
export LOCK_PREFIX="myapp:lock:"            # 锁前缀
```

### 代码配置

```go
import "time"

cfg := redCorn.Cfg{
    RedisCfg: redCorn.RedisCfg{
        Addr:     "your-redis:6379",
        Password: "yourpassword",
        DB:       0,
    },
    LockCfg: redCorn.LockCfg{
        Prefix: "custom:lock:",
        Expiry: 60 * time.Second,
    },
}
```

## 🔧 环境变量配置

虽然代码中没有内置的 `LoadConfig()` 函数，但你可以轻松地自己实现：

```go
import (
    "os"
    "strconv"
    "time"
)

func LoadConfig() redCorn.Cfg {
    // Redis配置
    redisDB := 0
    if db := os.Getenv("REDIS_DB"); db != "" {
        if parsedDB, err := strconv.Atoi(db); err == nil {
            redisDB = parsedDB
        }
    }
    
    // 锁过期时间
    lockExpiry := 30 * time.Second
    if expiry := os.Getenv("LOCK_EXPIRY"); expiry != "" {
        if parsedExpiry, err := time.ParseDuration(expiry); err == nil {
            lockExpiry = parsedExpiry
        }
    }
    
    return redCorn.Cfg{
        RedisCfg: redCorn.RedisCfg{
            Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
            Password: getEnv("REDIS_PASSWORD", ""),
            DB:       redisDB,
        },
        LockCfg: redCorn.LockCfg{
            Prefix: getEnv("LOCK_PREFIX", "redcorn:lock:"),
            Expiry: lockExpiry,
        },
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

## 📋 API 参考

### 核心结构

```go
// 配置结构体
type Cfg struct {
    RedisCfg RedisCfg
    LockCfg  LockCfg
    Logger   Logger // 自定义日志器，可选
}

type RedisCfg struct {
    Addr     string
    Password string
    DB       int
}

type LockCfg struct {
    Expiry time.Duration
    Prefix string
}

// 分布式任务管理器
type DistributedTaskManager struct {
    // 内部字段
}

// 任务调度器
type TaskScheduler struct {
    // 内部字段
}

// 任务调度信息
type TaskSchedule struct {
    Task func()
    Cron string
}
```

### 主要方法

```go
// 创建任务管理器
func NewDistributedTaskManager(cfg Cfg) (*DistributedTaskManager, error)

// 添加单个任务
func (dtm *DistributedTaskManager) AddTask(name, cron string, task func()) error

// 批量添加任务
func (dtm *DistributedTaskManager) AddScheduler(scheduler *TaskScheduler) error

// 启动任务管理器
func (dtm *DistributedTaskManager) Start()

// 停止任务管理器
func (dtm *DistributedTaskManager) Stop()

// 获取Redis客户端（供外部使用）
func (dtm *DistributedTaskManager) GetRedisClient() *goredislib.Client

// 获取上下文（供外部使用）
func (dtm *DistributedTaskManager) GetContext() context.Context

// 创建任务调度器
func NewTaskScheduler() *TaskScheduler

// 注册任务
func (ts *TaskScheduler) Register(name, cron string, task func())

// 获取任务
func (ts *TaskScheduler) Get(name string) (TaskSchedule, bool)

// 获取所有任务
func (ts *TaskScheduler) GetAll() map[string]TaskSchedule
```

## 📝 Cron 表达式

支持标准 6 位 Cron 表达式（包含秒），使用 [robfig/cron](https://github.com/robfig/cron) 库：

| 表达式 | 含义 |
|--------|------|
| `*/10 * * * * *` | 每 10 秒 |
| `0 */5 * * * *` | 每 5 分钟 |
| `0 0 * * * *` | 每小时 |
| `0 0 2 * * *` | 每天 2 点 |
| `0 30 9 * * 1-5` | 工作日 9:30 |

**注意：** 不支持 `?` 符号，请使用标准的 `*` 符号。

## 🔄 重试机制

RedCorn 内置智能重试机制：

- **随机延迟**：避免多个节点同时重试
- **可配置重试次数**：通过配置控制最大重试次数
- **指数退避**：可选的指数退避策略

## 🛠️ 自定义日志

实现 `Logger` 接口来自定义日志：

```go
type Logger interface {
    Debug(args ...interface{})
    Info(args ...interface{})
    Warn(args ...interface{})
    Error(args ...interface{})
    Fatal(args ...interface{})
}

// 使用自定义日志器
cfg := redCorn.Cfg{
    RedisCfg: redisCfg,
    LockCfg:  lockCfg,
    Logger:   &MyLogger{}, // 你的日志器
}
```

## 🧪 示例项目

查看 [example](example/) 目录获取完整示例：

- `main.go` - 基础使用示例

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 🔗 相关链接

- [go-redis](https://github.com/redis/go-redis) - Redis 客户端
- [redsync](https://github.com/go-redsync/redsync) - Redis 分布式锁
- [robfig/cron](https://github.com/robfig/cron) - Cron 表达式解析器