package service

import (
	"math"
	"time"
)

type Task func(loopNo int) bool

type Result struct {
	TaskNo      int
	AverageTime int
	SuccessNo   int
	MaxTime     int
	MinTime     int
}

func Benchmark(amount int, task Task) *Result {
	taskNo := amount
	successTaskNo := 0
	totalTime := 0
	maxTime := 0
	minTime := math.MaxInt32

	for index := 0; index < taskNo; index++ {
		start := time.Now()

		if !task(index) {
			continue
		}

		elapse := time.Now()
		elapseT := int(elapse.UnixNano() - start.UnixNano())

		if maxTime < elapseT {
			maxTime = elapseT
		}

		if minTime > elapseT {
			minTime = elapseT
		}

		totalTime = totalTime + elapseT

		successTaskNo++
	}

	average := totalTime / successTaskNo

	return &Result{
		TaskNo:      amount,
		SuccessNo:   successTaskNo,
		AverageTime: average,
		MaxTime:     maxTime,
		MinTime:     minTime,
	}
}
