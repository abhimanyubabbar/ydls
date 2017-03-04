// Package leaktest checks for leaks of goroutines, file descriptors,
// child processes and temp files.
//
// File descriptors check will not detect if a fd number existing before the
// check start and then is reused by leaky code.
//
// Temp files check requires that $TMPDIR is used (os.TempDir or
// ioutil.Temp* with empty dir argument).
//
// It is important that the returned check function is always called as it
// also restores TMPDIR. Use defer or func() { defer ... }().
//
package leaktest

// build +linux

// TODO: make it work on osx
// TODO: possible to use procfs fd inode, ctime etc?

// goroutine leaktest code from:
// https://github.com/fortytw2/leaktest
// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

func stringSetMinus(as []string, bs []string) []string {
	nmap := map[string]struct{}{}
	for _, a := range as {
		nmap[a] = struct{}{}
	}

	for _, b := range bs {
		delete(nmap, b)
	}

	ns := []string{}
	for n := range nmap {
		ns = append(ns, n)
	}

	return ns
}

func intSetMinus(as []int, bs []int) []int {
	nmap := map[int]struct{}{}
	for _, a := range as {
		nmap[a] = struct{}{}
	}

	for _, b := range bs {
		delete(nmap, b)
	}

	ns := []int{}
	for n := range nmap {
		ns = append(ns, n)
	}

	return ns
}

// interestingGoroutines returns all goroutines we care about for the purpose
// of leak checking. It excludes testing or runtime ones.
func interestingGoroutines() (gs []string) {
	buf := make([]byte, 2<<20)
	buf = buf[:runtime.Stack(buf, true)]
	for _, g := range strings.Split(string(buf), "\n\n") {
		sl := strings.SplitN(g, "\n", 2)
		if len(sl) != 2 {
			continue
		}
		stack := strings.TrimSpace(sl[1])
		if strings.HasPrefix(stack, "testing.RunTests") {
			continue
		}

		if stack == "" ||
			// Below are the stacks ignored by the upstream leaktest code.
			strings.Contains(stack, "testing.Main(") ||
			strings.Contains(stack, "testing.(*T).Run(") ||
			strings.Contains(stack, "runtime.goexit") ||
			strings.Contains(stack, "created by runtime.gc") ||
			strings.Contains(stack, "interestingGoroutines") ||
			strings.Contains(stack, "runtime.MHeap_Scavenger") ||
			strings.Contains(stack, "signal.signal_recv") ||
			strings.Contains(stack, "sigterm.handler") ||
			strings.Contains(stack, "runtime_mcall") ||
			strings.Contains(stack, "goroutine in C code") {
			continue
		}
		gs = append(gs, strings.TrimSpace(g))
	}
	sort.Strings(gs)
	return
}

// TODO: could be done by probing fds with some syscall?
func fdsForCurrentProcess() ([]int, error) {
	fdFiles, err := ioutil.ReadDir("/dev/fd")
	if err != nil {
		return nil, err
	}
	fds := []int{}
	for _, fdFile := range fdFiles {
		fd, _ := strconv.Atoi(fdFile.Name())
		fds = append(fds, fd)
	}

	return fds, nil
}

type stat struct {
	name string
	ppid int
}

func readProcStatForPid(pid int) (stat, error) {
	statBuf, err := ioutil.ReadFile(fmt.Sprintf("/proc/%s/stat", strconv.Itoa(pid)))
	if err != nil {
		return stat{}, err
	}

	statParts := strings.Split(string(statBuf), " ")
	if len(statParts) < 4 {
		return stat{}, fmt.Errorf("failed to split stat for %d", pid)
	}

	ppid, _ := strconv.Atoi(statParts[3])
	return stat{
		name: statParts[1],
		ppid: ppid,
	}, nil
}

// childsForPid number of child processes for pid
// can't be done atomically so only use in controlled environments
func childsForPid(pid int) ([]int, error) {
	procFiles, err := ioutil.ReadDir("/proc")
	if err != nil {
		return nil, err
	}
	// build pid => parent pid map
	pidPpid := map[int]int{}
	for _, procFile := range procFiles {
		procPid, err := strconv.Atoi(procFile.Name())
		if err != nil {
			continue
		}

		stat, err := readProcStatForPid(procPid)
		if err != nil {
			continue
		}

		pidPpid[procPid] = stat.ppid
	}

	childs := []int{}
	var collectHelper func(pid int)
	collectHelper = func(pid int) {
		for cpid, ppid := range pidPpid {
			if ppid == pid {
				childs = append(childs, cpid)
				collectHelper(cpid)
			}
		}
	}

	collectHelper(pid)

	return childs, nil
}

// ErrorReporter is a tiny subset of a testing.TB to make testing not such a
// massive pain
type ErrorReporter interface {
	Errorf(format string, args ...interface{})
	Fatal(...interface{})
}

// Check check for leaked fds and child processes
// use defer osleaktest.Check(t)() first in test function
func Check(t ErrorReporter) func() {
	goroutinesBefore := interestingGoroutines()

	fdsBefore, fdsBeforeErr := fdsForCurrentProcess()
	if fdsBeforeErr != nil {
		t.Fatal(fdsBeforeErr)
	}
	childsBefore, childsBeforeErr := childsForPid(os.Getpid())
	if childsBeforeErr != nil {
		t.Fatal(childsBeforeErr)
	}

	testTempDir, testTempDirErr := ioutil.TempDir("", "leaktest")
	if testTempDirErr != nil {
		t.Fatal(testTempDirErr)
	}
	origTMPDIR := os.Getenv("TMPDIR")
	os.Setenv("TMPDIR", testTempDir)

	return func() {
		defer os.Setenv("TMPDIR", origTMPDIR)
		defer os.RemoveAll(testTempDir)

		fdsAfter, fdsAfterErr := fdsForCurrentProcess()
		if fdsAfterErr != nil {
			t.Fatal(fdsAfterErr)
		}
		childsAfter, childsAfterErr := childsForPid(os.Getpid())
		if childsAfterErr != nil {
			t.Fatal(childsAfterErr)
		}
		fdsLeaked := intSetMinus(fdsAfter, fdsBefore)
		if len(fdsLeaked) != 0 {
			t.Errorf("Leaked fds: %v", fdsLeaked)
		}

		childsLeaked := intSetMinus(childsAfter, childsBefore)
		if len(childsLeaked) != 0 {
			fancyPids := []string{}
			for _, pid := range childsLeaked {
				stat, err := readProcStatForPid(pid)
				if err == nil {
					fancyPids = append(fancyPids, fmt.Sprintf("%d %s", pid, stat.name))
				} else {
					fancyPids = append(fancyPids, strconv.Itoa(pid))
				}
			}
			t.Errorf("Leaked child processes: %v", fancyPids)
		}

		// Loop, waiting for goroutines to shut down.
		// Wait up to 5 seconds, but finish as quickly as possible.
		deadline := time.Now().Add(5 * time.Second)
		for {
			goroutinesLeaked := stringSetMinus(interestingGoroutines(), goroutinesBefore)
			if len(goroutinesLeaked) == 0 {
				break
			}
			if time.Now().Before(deadline) {
				time.Sleep(50 * time.Millisecond)
				continue
			}
			t.Errorf("Leaked goroutines: %v", goroutinesLeaked)
			break
		}

		leakedTempFiles := []string{}
		filepath.Walk(testTempDir, func(path string, info os.FileInfo, err error) error {
			if path == testTempDir {
				return nil
			}
			leakedTempFiles = append(leakedTempFiles, path)
			return nil
		})
		if len(leakedTempFiles) != 0 {
			t.Errorf("Leaked temp files: %v", leakedTempFiles)
		}
	}
}
