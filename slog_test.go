// Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors
// The use of this source code is governed by a MIT style license found in the LICENSE file

package slog_test

import (
	"testing"
	"github.com/ventu-io/slog"
	"github.com/ventu-io/slf"
)

func TestLogger_compliesWithInterface_success(t *testing.T) {
	var _ slf.Logger = slog.New()
}

