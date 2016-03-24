// Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors
// The use of this source code is governed by a MIT style license found in the LICENSE file

package text_test

import (
	"errors"
	"github.com/ventu-io/slf"
	"github.com/ventu-io/slog"
	"github.com/ventu-io/slog/text"
	"testing"
)

type stringwriter struct {
	res string
	err error
}

func (sw *stringwriter) Write(p []byte) (n int, err error) {
	sw.res = sw.res + string(p)
	return len(p), sw.err
}

func TestPreliminary(t *testing.T) {
	lf := slog.New()
	h := text.New()
	wr := &stringwriter{}
	h.SetWriter(wr)
	lf.AddEntryHandler(h)
	lf.SetConcurrent(false)

	slf.Set(lf)
	slf.WithContext("test").WithField("A", 25).Info("done")
	slf.WithContext("test").WithField("A", 25).WithCaller(slf.CallerShort).WithError(errors.New("some error")).Warn("done")
	t.Log(wr.res)
}

type voidwriter struct {
	counter int
	max     int
	done    chan bool
}

func (vw *voidwriter) Write(p []byte) (int, error) {
	vw.counter++
	if vw.counter >= vw.max {
		vw.done <- true
	}
	return len(p), nil
}

func log250k(logger slf.Logger) {
	for i := 0; i < 250000; i++ {
		logger.Infof("i=%v", i)
	}
}

func TestPerformance(t *testing.T) {
	f := slog.New()
	h := text.New()
	wr := &voidwriter{0, 1000000, make(chan bool)}
	h.SetWriter(wr)
	f.AddEntryHandler(h)
	slf.Set(f)
	logger0 := slf.WithContext("ctx1").WithField("A", 25)
	logger1 := logger0.WithField("B", 26)
	logger2 := logger0.WithError(errors.New("err"))
	logger3 := slf.WithContext("ctx2").WithCaller(slf.CallerShort)
	go log250k(logger0)
	go log250k(logger1)
	go log250k(logger2)
	go log250k(logger3)
	<-wr.done
}
