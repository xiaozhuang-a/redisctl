package utils

import (
	"sync"
)

type Task func() error

func TaskRunner(maxConcurrency int, tasks []Task) []error {
	var errChan = make(chan error, len(tasks))
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrency)
	for _, task := range tasks {
		wg.Add(1)
		sem <- struct{}{}
		go func(t Task) {
			defer func() {
				wg.Done()
				<-sem
			}()
			err := t()
			if err != nil {
				errChan <- err
			}
		}(task)
	}
	wg.Wait()
	close(errChan)
	var errList []error
	for err := range errChan {
		errList = append(errList, err)
	}
	return errList
}

type TaskWithResult[T interface{}] func() ([]T, error)

// TaskRunnerWithResult 并发执行带有返回值的任务，且返回值的数量不固定
func TaskRunnerWithResult[T interface{}](maxConcurrency int, tasks []TaskWithResult[T]) ([]T, []error) {
	var (
		errChan = make(chan error, len(tasks))
		wg      sync.WaitGroup
		sem     = make(chan struct{}, maxConcurrency)
		result  = NewSafeSlice[T]()
	)

	for _, task := range tasks {
		wg.Add(1)
		sem <- struct{}{}
		go func(t TaskWithResult[T]) {
			defer func() {
				wg.Done()
				<-sem
			}()
			data, err := t()
			if err != nil {
				errChan <- err
			} else {
				result.Append(data...)
			}
		}(task)
	}
	wg.Wait()
	close(errChan)
	var errList []error
	for err := range errChan {
		errList = append(errList, err)
	}
	return result.All(), errList
}
