// Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors
// The use of this source code is governed by a MIT style license found in the LICENSE file

package text_test

import (
	"errors"
	"github.com/ventu-io/slf"
	"github.com/ventu-io/slog"
	"github.com/ventu-io/slog/text"
	"testing"
	"time"
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
	f := slog.New()
	h := text.New()
	wr := &stringwriter{}
	h.SetWriter(wr)
	f.AddEntryHandler(h)
	slf.Set(f)
	slf.WithContext("test").WithField("A", 25).Info("done")
	slf.WithContext("test").WithField("A", 25).WithCaller(slf.CallerShort).WithError(errors.New("some error")).Warn("done")
	time.Sleep(time.Second)
	t.Log(wr.res)
}
