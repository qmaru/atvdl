package utils

import (
	"sync"
)

// TaskConfig 任务配置
//
//	Control 控制状态 Add() Done() Wait()
//	Mutex 线程安全锁 Lock() Unlock()
//	Workers 线程数
//	Results 任务执行结果
//	Counts 任务总数
type TaskConfig struct {
	Control sync.WaitGroup
	Counts  int
	Mutex   sync.Mutex
	Results chan interface{}
	Workers chan interface{}
}

// TaskBoard 任务公告栏
//
//	runner 执行任务的函数
//	tasks 任务列表
//	worker 处理任务的线程数
func TaskBoard(runner func(task interface{}) (interface{}, error), tasks []string, workers int) ([]interface{}, error) {
	var data []interface{}

	// 任务配置
	taskCfg := new(TaskConfig)
	// 统计任务数量
	taskCfg.Counts = len(tasks)
	// 分配线程
	taskCfg.Workers = make(chan interface{}, workers)
	// 任务结果缓存
	taskCfg.Results = make(chan interface{}, 1024)
	// 添加任务
	taskCfg.Control.Add(taskCfg.Counts)
	// 任务分配
	for i := 0; i < taskCfg.Counts; i++ {
		task := tasks[i]
		// 领取任务
		taskCfg.Workers <- task
		go func() {
			defer taskCfg.Control.Done()
			// 执行任务
			result, _ := runner(task)
			taskCfg.Results <- result
			<-taskCfg.Workers
		}()
	}

	go func() {
		for taskResult := range taskCfg.Results {
			taskCfg.Mutex.Lock()
			data = append(data, taskResult)
			taskCfg.Mutex.Unlock()
		}
	}()

	taskCfg.Control.Wait()
	return data, nil
}
