package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kzdgt/redCorn"
)

func main() {
	log.Println("Starting Enhanced RedCorn Example...")

	// 加载配置
	config := redCorn.Cfg{}
	config.RedisCfg.Addrs = []string{"localhost:6379"}
	config.RedisCfg.Password = "Qnzs@123"
	config.LockCfg.Prefix = "myapp:lock:"
	config.LockCfg.Expiry = 60 * time.Second

	// 创建分布式任务管理器
	dtm, err := redCorn.NewDistributedTaskManager(config)
	if err != nil {
		log.Fatalf("Failed to create distributed task manager: %v", err)
	}

	// 方式1：集中式管理
	scheduler := redCorn.NewTaskScheduler()

	// 在一个地方集中定义所有任务和调度信息
	scheduler.Register("health-check", "0/10 * * * * ? ", func() {
		log.Println("=== Health Check ===")
		time.Sleep(1 * time.Second)
		log.Println("Health check completed")
	})

	scheduler.Register("data-sync", "0/10 * * * * ? ", func() {
		log.Println("=== Data Sync ===")
		time.Sleep(3 * time.Second)
		log.Println("Data sync completed")
	})

	scheduler.Register("email-sender", "0/10 * * * * ? ", func() {
		log.Println("=== Email Sender ===")
		time.Sleep(2 * time.Second)
		log.Println("Email sender completed")
	})

	// 一次性添加所有任务
	if err := dtm.AddScheduler(scheduler); err != nil {
		log.Fatalf("Failed to add scheduler: %v", err)
	}

	// 方式2：灵活性补充 - 单独添加任务
	err = dtm.AddTask("simple-job", "0/10 * * * * ? ", func() {
		log.Println("=== Simple Job ===")
		time.Sleep(1 * time.Second)
		log.Println("Simple job completed")
	})
	if err != nil {
		log.Fatalf("Failed to add task: %v", err)
	}

	// 启动任务管理器
	dtm.Start()

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	dtm.Stop()
	log.Println("Enhanced RedCorn example stopped")
}
