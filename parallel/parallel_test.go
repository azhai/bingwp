package parallel

import (
	"fmt"
	"testing"
	"time"
)

func SetUpAction(id int) {
	time.Sleep(time.Duration(id*100) * time.Millisecond)
	fmt.Println("SetUp", id)
}

func LoopAction(id int, start, stop int64) {
	fmt.Println("Loop", id, start, stop)
}

func CreateWork(qChan chan bool, msecs int) *AlignWork {
	task1 := NewStageTask(SetUpAction, false)
	task2 := NewRangeTask(LoopAction, qChan, 1)
	task2.EachGap = msecs
	task2.Config = func(id int) (int64, int64) {
		var stop = int64(id)*100 + 1
		return stop - 100, stop
	}
	return new(AlignWork).Then(task1).Then(task2)
}

func RunScheduler(count, runSecs, gapMsecs int) {
	elapse := time.Duration(runSecs) * time.Second
	sch := NewScheduler(elapse, true)
	work := CreateWork(sch.QuitChan, gapMsecs)
	sch.Run(count, work)
}

func BenchmarRun(b *testing.B) {
	RunScheduler(b.N, 0, 1)
}

func TestRun(t *testing.T) {
	RunScheduler(10, 15, 50)
}

func TestRange1(t *testing.T) {
	start, end, step := 1, 88, 10
	fmt.Printf("from %d to %d:\n", start, end)
	temp := start + step
	for temp < end {
		fmt.Println(start, "->", temp-1)
		start, temp = temp, temp+step
	}
	fmt.Println(start, "->", end)
}

func TestRange2(t *testing.T) {
	num, step := 8, int64(10) // 8个并发，各自负责10个数
	fmt.Printf("from 1 to %d:\n", int64(num)*step)
	action := func(id int, start, stop int64) {
		fmt.Println(id, ":", start, "->", stop-1)
	}
	sch := NewScheduler(0, true)
	task := NewRangeTask(action, sch.QuitChan, step)
	task.NewConfig(num, 1, 88, true)
	sch.RunTask(num, task)
}
