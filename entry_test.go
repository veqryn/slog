// Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors
// The use of this source code is governed by a MIT style license found in the LICENSE file

package slog_test

import (
	"github.com/ventu-io/slf"
	"github.com/ventu-io/slog"
	"testing"
)

func TestLogger_conformsInterface_success(t *testing.T) {
	var _ slf.Logger = slog.New()
}
