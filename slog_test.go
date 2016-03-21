// Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors
// The use of this source code is governed by a MIT style license found in the LICENSE file

package slog_test

import (
	"github.com/ventu-io/slf"
	"github.com/ventu-io/slog"
	"testing"
)

func TestLogger_initialized_success(t *testing.T) {
	f := slog.New()
	if len(f.Contexts()) != 0 {
		t.Error("expected not contexts")
	}
}

func TestLogger_withContext_success(t *testing.T) {
	f := slog.New()
	logger0 := f.WithContext("test0")
	logger1 := f.WithContext("test0")
	logger2 := f.WithContext("test2")
	if logger0 != f.Contexts()["test0"] {
		t.Error("expected logger for test0")
	}
	if logger1 != f.Contexts()["test0"] {
		t.Error("expected logger for test0")
	}
	if logger2 != f.Contexts()["test2"] {
		t.Error("expected logger for test2")
	}
	if len(f.Contexts()) != 2 {
		t.Errorf("expected 2 contexts, found %v", len(f.Contexts()))
	}
}

func TestLogger_setLevelWithoutContext_resetsExistingAndRoot_success(t *testing.T) {
	th := &testhandler{done: make(chan bool)}
	f := slog.New()
	f.AddEntryHandler(th)
	logger := f.WithContext("test0")
	logger.Debug("debug1")
	logger.Info("info1")
	<-th.done
	if len(th.entries) != 1 || th.entries[0].Message() != "info1" {
		t.Errorf("incorrect log entries found, %v", th.entries)
	}
	th.entries = nil
	// resets all, of the above logger and for the newly created ones
	f.SetLevel(slf.LevelDebug)
	// this logger is affected
	logger.Debug("debug1")
	<-th.done
	logger.Info("info1")
	<-th.done
	if len(th.entries) != 2 || th.entries[0].Message() != "debug1" || th.entries[1].Message() != "info1" {
		t.Errorf("incorrect log entries found, %v", th.entries)
	}
	th.entries = nil
	// this logger is also affected
	logger = f.WithContext("test1")
	logger.Debug("debug1")
	<-th.done
	logger.Info("info1")
	<-th.done
	if len(th.entries) != 2 || th.entries[0].Message() != "debug1" || th.entries[1].Message() != "info1" {
		t.Errorf("incorrect log entries found, %v", th.entries)
	}
}

func TestLogger_setLevelWithContexts_success(t *testing.T) {
	th := &testhandler{done: make(chan bool)}
	f := slog.New()
	f.AddEntryHandler(th)
	logger1 := f.WithContext("test1")
	logger2 := f.WithContext("test2")
	f.SetLevel(slf.LevelDebug, "test1", "root")
	logger1.Debug("debug1")
	<-th.done
	logger1.Info("info1")
	<-th.done
	logger2.Debug("debug2")
	logger2.Info("info2")
	<-th.done
	logger3 := f.WithContext("test3")
	logger3.Debug("debug3")
	<-th.done
	logger3.Info("info3")
	<-th.done
	if len(th.entries) != 5 || th.entries[0].Message() != "debug1" ||
		th.entries[1].Message() != "info1" || th.entries[2].Message() != "info2" ||
		th.entries[3].Message() != "debug3" || th.entries[4].Message() != "info3" {
		t.Errorf("incorrect log entries found, %v", th.entries)
	}
}

func TestLogger_addEntryHandlers_success(t *testing.T) {
	th0 := &testhandler{done: make(chan bool)}
	th1 := &testhandler{done: make(chan bool)}
	th2 := &testhandler{done: make(chan bool)}
	th3 := &testhandler{done: make(chan bool)}
	f := slog.New()
	f.AddEntryHandler(th0)
	f.SetEntryHandlers([]slog.EntryHandler{th1, th2}...)
	f.AddEntryHandler(th3)
	f.WithContext("test0").Info("test")
	<-th1.done
	<-th2.done
	<-th3.done
	if len(th0.done) != 0 {
		t.Error("th0 not expected")
	}
}
