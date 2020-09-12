package runner

import (
	"fmt"
	"strconv"
	"time"
)

type testResult struct {
	success   bool
	startTime time.Time
	endTime   time.Time
	err       error
}

type testResults map[string]testResult

func (r testResults) GetDuration(testID string) time.Duration {
	return r[testID].endTime.Sub(r[testID].startTime)
}

func (r testResults) GetSuccess(testID string) bool {
	return r[testID].success
}

func (r testResults) Start(testID string) {
	r[testID] = testResult{
		success:   false,
		startTime: time.Now(),
	}
}

func (r testResults) End(testID string, v bool, err error) {
	result := r[testID]
	result.success = v
	result.endTime = time.Now()
	result.err = err
	r[testID] = result
}

func (r testResults) PrintErrors() {
	if len(r) == 0 {
		return
	}

	fmt.Println("ERRORS:")

	maxLength := r.getMaxNameLength()

	for testName, result := range r {
		if result.err == nil {
			continue
		}

		fmt.Printf("%"+strconv.Itoa(maxLength)+"s: %v\n", testName, result.err)
	}
}

func (r testResults) GetFailed() testResults {
	failed := testResults{}

	for testName, result := range r {
		if result.err == nil {
			continue
		}

		failed[testName] = result
	}

	return failed
}

func (r testResults) getMaxNameLength() int {
	var max int

	for testName := range r {
		l := len(testName)
		if l > max {
			max = l
		}
	}

	return max
}
