package main

import (
	"sync"
	"sync/atomic"
	"time"
)

type completionBarrier interface {
	grabWork() bool
	jobDone()
	wait()
}

type countingCompletionBarrier struct {
	numReqs, reqsDone uint64
	doneCallback      func()
	wg                sync.WaitGroup
}

func newCountingCompletionBarrier(numReqs uint64, callback func()) completionBarrier {
	c := new(countingCompletionBarrier)
	c.reqsDone, c.numReqs = 0, numReqs
	c.doneCallback = callback
	c.wg.Add(int(numReqs))
	return completionBarrier(c)
}

func (c *countingCompletionBarrier) grabWork() bool {
	return atomic.AddUint64(&c.reqsDone, 1) <= c.numReqs
}

func (c *countingCompletionBarrier) jobDone() {
	c.doneCallback()
	c.wg.Done()
}

func (c *countingCompletionBarrier) wait() {
	c.wg.Wait()
}

type timedCompletionBarrier struct {
	wg           sync.WaitGroup
	tickCallback func()
	done         int64
}

func newTimedCompletionBarrier(parties int, duration time.Duration, callback func()) completionBarrier {
	c := new(timedCompletionBarrier)
	c.tickCallback = callback
	c.done = 0
	c.wg.Add(parties)
	go func() {
		secs := int(duration.Seconds())
		for i := 1; i <= secs; i++ {
			c.tickCallback()
			time.Sleep(1 * time.Second)
		}
		atomic.CompareAndSwapInt64(&c.done, 0, 1)
	}()
	return completionBarrier(c)
}

func (c *timedCompletionBarrier) grabWork() bool {
	done := atomic.LoadInt64(&c.done)
	if done == 1 {
		c.wg.Done()
	}
	return done == 0
}

func (c *timedCompletionBarrier) jobDone() {
}

func (c *timedCompletionBarrier) wait() {
	c.wg.Wait()
}
