package main

import (
	"fmt"
)

type BatchFunc func() (bool, error)

type BatchTask struct {
	Name string
	Func BatchFunc
}

type BatchRunner struct {
	tasks []BatchTask
}

func NewBatchRunner(tasks []BatchTask) *BatchRunner {
	return &BatchRunner{
		tasks: tasks,
	}
}

func (br *BatchRunner) Run() (bool, error) {
	fmt.Println("開始批次執行...")
	for i, task := range br.tasks {
		fmt.Printf("--- 正在執行任務: %s (序號: %d) ---\n", task.Name, i+1)

		success, err := task.Func()
		if err != nil {
			fmt.Printf("任務 '%s' 執行失敗，錯誤: %v\n", task.Name, err)
			return false, fmt.Errorf("任務 '%s' 執行失敗: %w", task.Name, err)
		}
		if !success {
			fmt.Printf("任務 '%s' 執行結果為失敗，停止批次執行。\n", task.Name)
			return false, fmt.Errorf("任務 '%s' 返回失敗狀態，批次中止", task.Name)
		}

		fmt.Printf("任務 '%s' 執行成功。\n\n", task.Name)
	}

	fmt.Println("所有批次任務執行完成。")
	return true, nil
}
