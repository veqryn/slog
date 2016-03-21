// Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors
// The use of this source code is governed by a MIT style license found in the LICENSE file

package slog_test

import "github.com/ventu-io/slog"

type testhandler struct {
	entries []slog.Entry
	done    chan bool
}

func (th *testhandler) Handle(entry slog.Entry) {
	th.entries = append(th.entries, entry)
	th.done <- true
}
