package utils

import (
	"github.com/sirupsen/logrus"
	"testing"
	"time"
)

func TestTaskRunnerWithResult(t *testing.T) {
	type N struct {
		Number int
	}

	tasks := make([]TaskWithResult[N], 0)
	for i := 1; i <= 10; i++ {
		tasks = append(tasks, func() ([]N, error) {
			time.Sleep(500 * time.Millisecond)
			return []N{{Number: i}}, nil
		})
	}
	st := time.Now()
	result, errs := TaskRunnerWithResult[N](5, tasks)
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	logrus.Infof("maxConcurrency: %d, cost: %v", 5, time.Since(st))
	st = time.Now()
	result, errs = TaskRunnerWithResult[N](10, tasks)
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	logrus.Infof("maxConcurrency: %d, cost: %v", 10, time.Since(st))
	logrus.Info(len(result))

}
