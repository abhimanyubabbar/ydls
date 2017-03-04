// +build linux

package leaktest

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"
)

// goroutine leaktest code from:
// https://github.com/fortytw2/leaktest
// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
type testReporter struct {
	failed bool
	msg    string
}

func (tr *testReporter) Errorf(format string, args ...interface{}) {
	tr.failed = true
	tr.msg = fmt.Sprintf(format, args)
}

func (tr *testReporter) Fatal(...interface{}) {
}

func tesFn(t *testing.T, inGorutine bool, shouldFail bool, fn func()) {
	checker := &testReporter{}
	checkFn := Check(checker)

	if inGorutine {
		go fn()
	} else {
		fn()
	}

	checkFn()

	if checker.failed != shouldFail {
		if shouldFail {
			t.Errorf("failed to detect leak")
		} else {
			t.Errorf("failed when there was no leak")
		}
	}
}

func testGoroutineLeak(t *testing.T, fn func()) {
	tesFn(t, true, true, fn)
}

func testLeak(t *testing.T, fn func()) {
	tesFn(t, false, true, fn)
}

func testNoLeak(t *testing.T, fn func()) {
	tesFn(t, false, false, fn)
}

func TestInfiniteForLoop(t *testing.T) {
	testGoroutineLeak(t, func() {
		for {
			time.Sleep(time.Second)
		}
	})
}

func TestSelectUnreferencedChannel(t *testing.T) {
	testGoroutineLeak(t, func() {
		c := make(chan struct{}, 0)
		select {
		case <-c:
		}
	})
}

func TestBlockSelectUnreferencedChannel(t *testing.T) {
	testGoroutineLeak(t, func() {
		c := make(chan struct{}, 0)
		c2 := make(chan struct{}, 0)
		select {
		case <-c:
		case c2 <- struct{}{}:
		}
	})
}

func TestBlockWaitMutexUnreferenced(t *testing.T) {
	testGoroutineLeak(t, func() {
		var mu sync.Mutex
		mu.Lock()
		mu.Lock()
	})
}

func TestBlockWaitRWMutexUnreferenced(t *testing.T) {
	testGoroutineLeak(t, func() {
		var mu sync.RWMutex
		mu.RLock()
		mu.Lock()
	})
}

func TestBlockWaitMutexCondUnreferenced(t *testing.T) {
	testGoroutineLeak(t, func() {
		var mu sync.Mutex
		mu.Lock()
		c := sync.NewCond(&mu)
		c.Wait()
	})
}

func TestLeakFd(t *testing.T) {
	testLeak(t, func() {
		os.Pipe()
	})
}

func TestLeakFdBeforeClosed(t *testing.T) {
	r, w, _ := os.Pipe()
	testLeak(t, func() {
		os.Pipe()
		r.Close()
		w.Close()
	})
}

func TestLeakFdBeforeClosedNoLeak(t *testing.T) {
	r, w, _ := os.Pipe()
	testNoLeak(t, func() {
		r.Close()
		w.Close()
	})
}

func TestLeakChildProcessBeforeClosed(t *testing.T) {
	cmd1 := exec.Command("cat")
	cmd1.Start()
	testLeak(t, func() {
		cmd1.Process.Kill()
		cmd1.Wait()
		cmd2 := exec.Command("cat")
		cmd2.Start()
	})
}

func TestLeakChildProcessBeforeNoLeak(t *testing.T) {
	cmd1 := exec.Command("cat")
	cmd1.Start()
	testNoLeak(t, func() {
		cmd1.Process.Kill()
		cmd1.Wait()
	})
}

func TestLeakChildProcess(t *testing.T) {
	testLeak(t, func() {
		cmd := exec.Command("cat")
		cmd.Start()
	})
}

func TestLeakTempFile(t *testing.T) {
	testLeak(t, func() {
		ioutil.TempFile("", "testleak")
	})
}

func TestEmptyLeak(t *testing.T) {
	defer Check(t)()
	time.Sleep(time.Second)
}
