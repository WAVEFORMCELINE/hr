package main

import (
	"fmt"
	"sync"
	"time"
)

// Ttype represents a task with creation and finish times and result
type Ttype struct {
	id         int
	cT         string // время создания
	fT         string // время выполнения
	taskRESULT string
}

func main() {
	taskChan := make(chan Ttype)

	taskCreator := func(taskChan chan Ttype) {
		go func() {
			for {
				creationTime := time.Now().Format(time.RFC3339)
				if time.Now().Nanosecond()%2 > 0 { 
					creationTime = "Some error occurred"
				}
				task := Ttype{
					cT: creationTime,
					id: int(time.Now().Unix()),
				}
				taskChan <- task
			}
		}()
	}

	go taskCreator(taskChan)

	taskWorker := func(task Ttype) Ttype {
		parsedTime, err := time.Parse(time.RFC3339, task.cT)
		if err != nil || parsedTime.After(time.Now().Add(-20*time.Second)) {
			task.taskRESULT = "something went wrong"
		} else {
			task.taskRESULT = "task has been successful"
		}
		task.fT = time.Now().Format(time.RFC3339Nano)
		time.Sleep(time.Millisecond * 150) 
		return task
	}

	doneTasks := make(chan Ttype)
	undoneTasks := make(chan error)

	taskSorter := func(task Ttype) {
		if task.taskRESULT == "task has been successful" {
			doneTasks <- task
		} else {
			undoneTasks <- fmt.Errorf("Task id %d creation time %s, error %s", task.id, task.cT, task.taskRESULT)
		}
	}

	var wg sync.WaitGroup

	go func() {
		// получение тасков
		for task := range taskChan {
			wg.Add(1)
			go func(t Ttype) {
				defer wg.Done()
				processedTask := taskWorker(t)
				taskSorter(processedTask)
			}(task)
		}
	}()

	var resultMutex sync.Mutex
	result := make(map[int]Ttype)
	var errMutex sync.Mutex
	var errors []error

	go func() {
		for doneTask := range doneTasks {
			resultMutex.Lock()
			result[doneTask.id] = doneTask
			resultMutex.Unlock()
		}
	}()

	go func() {
		for undoneTask := range undoneTasks {
			errMutex.Lock()
			errors = append(errors, undoneTask)
			errMutex.Unlock()
		}
	}()

	time.Sleep(3 * time.Second)
	close(taskChan)
	wg.Wait()
	close(doneTasks)
	close(undoneTasks)

	fmt.Println("Errors:")
	for _, err := range errors {
		fmt.Println(err)
	}

	fmt.Println("Done tasks:")
	for _, task := range result {
		fmt.Println(task)
	}
}
