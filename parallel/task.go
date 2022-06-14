package parallel

import (
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/azhai/bingwp/utils"
)

// 间隔重复执行
func TickTime(ms int) (tick <-chan time.Time) {
	if ms > 0 {
		duration := time.Duration(ms) * time.Millisecond
		tick = time.Tick(duration)
	}
	return
}

// 延时，避免拥塞
func DelayTime(ms int) {
	if ms > 0 {
		duration := time.Duration(ms) * time.Millisecond
		time.Sleep(duration)
	}
}

// 随机延时
func DelayRand(ms int) {
	if ms > 0 {
		DelayTime(utils.RandInt(ms))
	}
}

// 任务接口
type ITask interface {
	IsRoutine() bool
	Process(id int)
}

// 阶段任务
type StageTask struct {
	isRoutine bool // 是否协程中执行
	process   func(id int)
}

func NewStageTask(process func(id int), isRoutine bool) *StageTask {
	return &StageTask{process: process, isRoutine: isRoutine}
}

func (t StageTask) IsRoutine() bool {
	return t.isRoutine
}

func (t *StageTask) Process(id int) {
	t.process(id)
}

// 区间任务
type RangeTask struct {
	QuitChan chan bool
	Action   func(id int, start, stop int64)
	Config   func(id int) (int64, int64) // id从1开始
	Step     int64                       // 步进，无限循环为0，一般为1即可
	EachGap  int                         // 每次的休息间隔，单位ms
	MaxDelay int                         // 启动的最大延迟，单位ms
}

func NewRangeTask(action func(id int, start, stop int64),
	quitChan chan bool, step int64) *RangeTask {
	return &RangeTask{Action: action, QuitChan: quitChan, Step: step}
}

func (t RangeTask) IsRoutine() bool {
	return true
}

// GetRange 计算起止范围，包含左边界 [start, stop)
func (t RangeTask) GetRange(id int) (start, stop int64) {
	if t.Config != nil {
		start, stop = t.Config(id)
	} else if t.Step > 1 {
		stop = int64(id) * t.Step
		start = stop - t.Step
	} else {
		stop = 1 // 默认为0和1
	}
	return
}

// NewConfig 根据起止范围生成 Config
func (t *RangeTask) NewConfig(n int, left, right int64, inclRight bool) {
	if inclRight { // 不包含右边界
		right = right + 1
	}
	var count, width int64
	if n < 1 {
		n = runtime.NumCPU()
	}
	if count = right - left; count < 1 || t.Step < 1 {
		return
	}
	if width = count / int64(n); width < 1 {
		width = 1
	} else { // 成为 t.Step 的整数倍
		width = (width + t.Step - 1) / t.Step * t.Step
	}
	t.Config = func(id int) (start, stop int64) {
		start = left + int64(id-1)*width
		if stop = start + width; stop > right {
			stop = right
		}
		return
	}
}

func (t *RangeTask) Process(id int) {
	// 计算启动和中间休眠时间
	var gapTime time.Duration
	if t.EachGap > 0 {
		gapTime = time.Duration(t.EachGap) * time.Millisecond
	}
	if t.MaxDelay > 0 {
		DelayRand(t.MaxDelay) // 随机延迟
	}

	temp := int64(0)
	start, stop := t.GetRange(id)
	// 执行循环
	for start < stop {
		select {
		default:
			if temp = start + t.Step; temp > stop {
				temp = stop
			}
			t.Action(id, start, temp)
			start = temp
			if gapTime > 0 {
				time.Sleep(gapTime)
			}
		case <-t.QuitChan:
			return
		}
	}
}

// 重复任务
type LoopTask struct {
	*RangeTask
}

func NewLoopTask(action func(id int, start, stop int64),
	quitChan chan bool) *LoopTask {
	return &LoopTask{NewRangeTask(action, quitChan, 0)}
}

// 先完成全部完成Task0，再开始Task1，然后Task2
type AlignWork struct {
	tasks []ITask
}

func (w AlignWork) Count() int {
	return len(w.tasks)
}

func (w AlignWork) GetTask(index int) ITask {
	if index < 0 || index >= w.Count() {
		return nil
	}
	return w.tasks[index]
}

func (w *AlignWork) Then(task ITask) *AlignWork {
	w.tasks = append(w.tasks, task)
	return w
}

// 调度器
type Scheduler struct {
	waiter   *sync.WaitGroup // 同步锁
	SignChan chan os.Signal
	QuitChan chan bool
	Begin    time.Time          // 开始时间
	Elapse   time.Duration      // 超时或经历时间
	IsCtrl   bool               // 捕获Ctrl+C等系统信号
	Finally  func(s *Scheduler) //最后执行的收尾工作
}

func NewScheduler(elapse time.Duration, isCtrl bool) *Scheduler {
	runtime.GOMAXPROCS(runtime.NumCPU()) // 最多使用N个核
	if elapse < 1 {
		elapse = 366 * 24 * time.Hour // 1年
	}
	return &Scheduler{
		waiter: new(sync.WaitGroup),
		Elapse: elapse,
		IsCtrl: isCtrl,
	}
}

func (s *Scheduler) SetFinally(finalWork func(s *Scheduler)) {
	s.Finally = finalWork
}

func (s *Scheduler) ExecTask(task ITask, count int) {
	for id := 1; id <= count; id++ {
		task.Process(id)
	}
}

func (s *Scheduler) GoExecTask(task ITask, count int) {
	s.waiter.Add(count)
	for id := 1; id <= count; id++ {
		go func(id int) {
			defer s.waiter.Done()
			task.Process(id)
		}(id)
	}
	runtime.Gosched()
	s.waiter.Wait()
}

func (s *Scheduler) Run(count int, work *AlignWork) {
	if count < 1 {
		count = runtime.NumCPU()
	}
	s.SetUp(count)
	for i := 0; i < work.Count(); i++ {
		task := work.GetTask(i)
		if task.IsRoutine() {
			s.GoExecTask(task, count)
		} else {
			s.ExecTask(task, count)
		}
	}
	s.TearDown()
}

func (s *Scheduler) RunTask(count int, tasks ...ITask) {
	work := &AlignWork{tasks: tasks}
	s.Run(count, work)
}

func (s *Scheduler) SetUp(count int) {
	s.Begin = time.Now()
	s.QuitChan = make(chan bool, count)
	s.SignChan = make(chan os.Signal, 1)
	signal.Notify(s.SignChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-s.SignChan: // 按Ctrl+C终止
			if s.IsCtrl {
				s.TearDown()
			}
			return
		case <-time.After(s.Elapse): // 超时终止
			s.TearDown()
			return
		}
	}()
}

func (s *Scheduler) TearDown() {
	s.Elapse = time.Since(s.Begin)
	if s.QuitChan != nil {
		for i := 0; i < cap(s.QuitChan); i++ {
			s.QuitChan <- true
		}
		close(s.QuitChan)
	}
	if s.Finally != nil {
		s.Finally(s)
	}
	// os.Exit(0)
}
