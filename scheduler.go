package redCorn

// TaskSchedule 任务调度定义
type TaskSchedule struct {
	Task func()
	Cron string
}

// TaskScheduler 任务调度器 - 集中管理任务和定时信息
type TaskScheduler struct {
	tasks map[string]TaskSchedule
}

// NewTaskScheduler 创建任务调度器
func NewTaskScheduler() *TaskScheduler {
	return &TaskScheduler{
		tasks: make(map[string]TaskSchedule),
	}
}

// Register 注册任务和定时信息
func (ts *TaskScheduler) Register(name string, cron string, task func()) {
	ts.tasks[name] = TaskSchedule{
		Task: task,
		Cron: cron,
	}
}

// Get 获取任务调度信息
func (ts *TaskScheduler) Get(name string) (TaskSchedule, bool) {
	schedule, exists := ts.tasks[name]
	return schedule, exists
}

// GetAll 获取所有任务调度信息
func (ts *TaskScheduler) GetAll() map[string]TaskSchedule {
	return ts.tasks
}
